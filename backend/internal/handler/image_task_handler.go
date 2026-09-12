package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	ctxkey "github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	pkghttputil "github.com/AsukaCC/EasySub2api/internal/pkg/httputil"
	"github.com/AsukaCC/EasySub2api/internal/pkg/logger"
	middleware2 "github.com/AsukaCC/EasySub2api/internal/server/middleware"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9" //nolint:depguard // queue message types are part of the Redis stream adapter contract
	"go.uber.org/zap"
)

type AsyncImageHandler struct {
	tasks       *service.ImageTaskService
	openAI      *OpenAIGatewayHandler
	execute     func(platform string, c *gin.Context)
	queue       imageTaskQueue
	apiKeys     *service.APIKeyService
	workers     int
	maxAttempts int
	claimIdle   time.Duration
	cancelMu    sync.Mutex
	cancel      map[string]context.CancelFunc
}

type imageTaskQueue interface {
	Ensure(context.Context) error
	Enqueue(context.Context, string) error
	Read(context.Context, string, time.Duration) ([]redis.XMessage, error)
	Claim(context.Context, string, time.Duration) ([]redis.XMessage, error)
	Ack(context.Context, string) error
}

func NewAsyncImageHandler(tasks *service.ImageTaskService, openAI *OpenAIGatewayHandler) *AsyncImageHandler {
	h := &AsyncImageHandler{tasks: tasks, openAI: openAI, cancel: make(map[string]context.CancelFunc)}
	h.execute = h.executeWithGateway
	return h
}

func NewAsyncImageHandlerWithQueue(tasks *service.ImageTaskService, openAI *OpenAIGatewayHandler, queue imageTaskQueue, apiKeys *service.APIKeyService, workers int, maxAttempts int, claimIdleSeconds int) *AsyncImageHandler {
	h := NewAsyncImageHandler(tasks, openAI)
	h.queue, h.apiKeys, h.workers, h.maxAttempts = queue, apiKeys, workers, maxAttempts
	if h.workers <= 0 {
		h.workers = 2
	}
	if h.maxAttempts < 1 {
		h.maxAttempts = 3
	}
	h.claimIdle = time.Duration(claimIdleSeconds) * time.Second
	if h.claimIdle <= 0 {
		h.claimIdle = 5 * time.Minute
	}
	if h.queue != nil {
		if err := h.queue.Ensure(context.Background()); err != nil {
			logger.L().Error("image_task.queue_init_failed", zap.Error(err))
		}
		go h.startWorkers()
	}
	return h
}

// enabled reports whether the async image task feature is available. Object
// storage is the enablement gate: without it the endpoints are fully disabled
// so that large base64 results never land in Redis.
func (h *AsyncImageHandler) enabled() bool {
	return h != nil && h.tasks != nil && h.tasks.Enabled()
}

// pollable reports whether task lookups can be served. It is deliberately weaker
// than enabled(): results already written to Redis stay readable after the
// feature is switched off, so an in-flight task is never stranded.
func (h *AsyncImageHandler) pollable() bool {
	return h != nil && h.tasks != nil && h.tasks.Pollable()
}

// Submit accepts the same payload as the synchronous Images endpoint and
// returns before the upstream image generation begins.
func (h *AsyncImageHandler) Submit(c *gin.Context) {
	if !h.enabled() {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "async image tasks are not enabled")
		return
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.UserID == "" || apiKey.ID == "" {
		imageTaskError(c, service.ErrImageTaskForbidden)
		return
	}
	platform := ""
	if apiKey.Group != nil {
		platform = apiKey.Group.Platform
	}
	if platform != service.PlatformOpenAI && platform != service.PlatformGrok {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "Images API is not supported for this platform")
		return
	}
	if !service.GroupAllowsImageGeneration(apiKey.Group) {
		imageTaskJSONError(c, http.StatusForbidden, "permission_error", service.ImageGenerationPermissionMessage())
		return
	}
	if h == nil || h.tasks == nil || h.execute == nil {
		imageTaskError(c, service.ErrImageTaskUnavailable)
		return
	}

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			imageTaskJSONError(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}
	if asyncImageRequestStreams(c.GetHeader("Content-Type"), body) {
		imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", "streaming image requests cannot be submitted as asynchronous tasks")
		return
	}
	if err := h.validateRequest(c, platform, body); err != nil {
		imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}
	if !h.checkSecurityAuditBeforeSubmit(c, apiKey, platform, body) {
		return
	}

	endpoint := "generations"
	if strings.Contains(c.Request.URL.Path, "/edits") {
		endpoint = "edits"
	}
	if h.queue != nil {
		store := h.tasks.ArtifactStore()
		if store == nil {
			imageTaskJSONError(c, http.StatusServiceUnavailable, "IMAGE_TASK_UNAVAILABLE", "async image object storage is unavailable")
			return
		}
		taskID := "imgtask_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		payloadKey := "image-tasks/" + taskID + "/request"
		if err := store.Put(c.Request.Context(), payloadKey, c.GetHeader("Content-Type"), body); err != nil {
			imageTaskJSONError(c, http.StatusServiceUnavailable, "IMAGE_TASK_UNAVAILABLE", "failed to store image task request")
			return
		}
		task, err := h.tasks.CreateQueued(c.Request.Context(), service.ImageTaskOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID}, platform, endpoint, c.GetHeader("Content-Type"), payloadKey)
		if err == nil {
			err = h.queue.Enqueue(c.Request.Context(), task.ID)
		}
		if err != nil {
			_ = store.Delete(context.Background(), payloadKey)
			imageTaskError(c, err)
			return
		}
		pollURL := imageTaskPollURL(c.Request.URL.Path, task.ID)
		c.Header("Cache-Control", "no-store")
		c.Header("Location", pollURL)
		c.Header("Retry-After", "3")
		c.JSON(http.StatusAccepted, gin.H{"id": task.ID, "task_id": task.TaskID, "object": task.Object, "status": task.Status, "created_at": task.CreatedAt, "expires_at": task.ExpiresAt, "poll_url": pollURL})
		return
	}

	taskCtx, recorder, cancel := newAsyncImageContext(c, body, h.tasks.ExecutionTimeout())
	task, err := h.tasks.Create(c.Request.Context(), service.ImageTaskOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID})
	if err != nil {
		cancel()
		imageTaskError(c, err)
		return
	}

	pollURL := imageTaskPollURL(c.Request.URL.Path, task.ID)
	c.Header("Cache-Control", "no-store")
	c.Header("Location", pollURL)
	c.Header("Retry-After", "3")
	c.JSON(http.StatusAccepted, gin.H{
		"id":         task.ID,
		"task_id":    task.TaskID,
		"object":     task.Object,
		"status":     task.Status,
		"created_at": task.CreatedAt,
		"expires_at": task.ExpiresAt,
		"poll_url":   pollURL,
	})

	go h.run(task.ID, platform, taskCtx, recorder, cancel)
}

func (h *AsyncImageHandler) Cancel(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		imageTaskError(c, service.ErrImageTaskForbidden)
		return
	}
	task, err := h.tasks.Cancel(c.Request.Context(), service.ImageTaskOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID}, c.Param("task_id"))
	if err != nil {
		imageTaskError(c, err)
		return
	}
	h.cancelMu.Lock()
	if cancel := h.cancel[task.ID]; cancel != nil {
		cancel()
	}
	delete(h.cancel, task.ID)
	h.cancelMu.Unlock()
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, task)
}

func (h *AsyncImageHandler) startWorkers() {
	if err := h.queue.Ensure(context.Background()); err != nil {
		logger.L().Error("image_task.queue_init_failed", zap.Error(err))
		return
	}
	for i := 0; i < h.workers; i++ {
		go h.workerLoop(fmt.Sprintf("image-worker-%d", i))
	}
}

func (h *AsyncImageHandler) workerLoop(consumer string) {
	for {
		messages, err := h.queue.Claim(context.Background(), consumer, h.claimIdle)
		if err == nil && len(messages) == 0 {
			messages, err = h.queue.Read(context.Background(), consumer, 5*time.Second)
		}
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, msg := range messages {
			taskID, _ := msg.Values["task_id"].(string)
			if taskID == "" {
				_ = h.queue.Ack(context.Background(), msg.ID)
				continue
			}
			h.processQueuedTask(taskID)
			_ = h.queue.Ack(context.Background(), msg.ID)
		}
	}
}

func (h *AsyncImageHandler) processQueuedTask(taskID string) {
	if h.apiKeys == nil {
		_ = h.tasks.Fail(context.Background(), taskID, http.StatusServiceUnavailable, imageTaskErrorPayload("api_error", "image task worker is unavailable"))
		return
	}
	record, err := h.tasks.Record(context.Background(), taskID)
	if err != nil {
		return
	}
	record.Attempt++
	if err := h.tasks.SaveRecord(context.Background(), record); err != nil {
		return
	}
	apiKey, err := h.apiKeys.GetByID(context.Background(), record.APIKeyID)
	if err != nil || apiKey == nil {
		_ = h.tasks.Fail(context.Background(), taskID, http.StatusUnauthorized, imageTaskErrorPayload("authentication_error", "API key is no longer available"))
		return
	}
	store := h.tasks.ArtifactStore()
	if store == nil {
		_ = h.tasks.Fail(context.Background(), taskID, http.StatusServiceUnavailable, imageTaskErrorPayload("api_error", "image task object storage is unavailable"))
		return
	}
	body, contentType, err := store.Get(context.Background(), record.PayloadKey)
	if err != nil {
		_ = h.tasks.Fail(context.Background(), taskID, http.StatusBadGateway, imageTaskErrorPayload("api_error", "failed to read image task request"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.tasks.ExecutionTimeout())
	h.cancelMu.Lock()
	h.cancel[taskID] = cancel
	h.cancelMu.Unlock()
	defer func() {
		cancel()
		h.cancelMu.Lock()
		delete(h.cancel, taskID)
		h.cancelMu.Unlock()
		_ = store.Delete(context.Background(), record.PayloadKey)
	}()
	ctx = context.WithValue(ctx, ctxkey.UserID, apiKey.UserID)
	ctx = context.WithValue(ctx, ctxkey.APIKeyGroupIDs, apiKey.BoundGroupIDs())
	req := httptest.NewRequest(http.MethodPost, "/v1/images/"+record.Endpoint, bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	writer := httptest.NewRecorder()
	taskCtx, _ := gin.CreateTestContext(writer)
	taskCtx.Request = req
	taskCtx.Set(string(middleware2.ContextKeyAPIKey), apiKey)
	taskCtx.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: apiKey.UserID})
	if apiKey.User != nil {
		taskCtx.Set(string(middleware2.ContextKeyUserRole), apiKey.User.Role)
	}
	h.execute(record.Platform, taskCtx)
	response := bytes.TrimSpace(writer.Body.Bytes())
	code := writer.Code
	if code == 0 {
		code = http.StatusOK
	}
	if ctx.Err() != nil && len(response) == 0 {
		_ = h.tasks.Fail(context.Background(), taskID, http.StatusGatewayTimeout, imageTaskErrorPayload("timeout_error", "image generation task timed out"))
		return
	}
	if code >= 200 && code < 300 && json.Valid(response) {
		_ = h.tasks.Complete(context.Background(), taskID, code, response)
		return
	}
	if code >= 500 && record.Attempt < h.maxAttempts {
		time.Sleep(time.Duration(record.Attempt*record.Attempt) * 5 * time.Second)
		if enqueueErr := h.queue.Enqueue(context.Background(), taskID); enqueueErr != nil {
			_ = h.tasks.Fail(context.Background(), taskID, code, extractImageTaskError(response))
		}
		return
	}
	_ = h.tasks.Fail(context.Background(), taskID, code, extractImageTaskError(response))
}

func (h *AsyncImageHandler) checkSecurityAuditBeforeSubmit(c *gin.Context, apiKey *service.APIKey, platform string, body []byte) bool {
	if h == nil || h.openAI == nil {
		return true
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		imageTaskJSONError(c, http.StatusInternalServerError, "api_error", "User context not found")
		return false
	}
	model := ""
	moderationBody := body
	if platform == service.PlatformGrok {
		parsed := service.ParseGrokMediaRequest(c.GetHeader("Content-Type"), body)
		model, moderationBody = parsed.Model, parsed.ModerationBody()
	} else if h.openAI.gatewayService != nil {
		parsed, err := h.openAI.gatewayService.ParseOpenAIImagesRequest(c, body)
		if err != nil {
			imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
			return false
		}
		model, moderationBody = parsed.Model, parsed.ModerationBody()
	}
	if len(moderationBody) == 0 {
		c.Set(securityAuditCompletedContextKey, true)
		return true
	}
	reqLog := requestLogger(c, "handler.async_image.security_audit",
		zap.String("user_id", subject.UserID), zap.String("api_key_id", apiKey.ID), zap.String("model", model))
	decision := h.openAI.checkSecurityAudit(c, reqLog, apiKey, subject, service.ContentModerationProtocolOpenAIImages, model, moderationBody)
	if decision != nil && !decision.AllowNextStage {
		h.openAI.openAISecurityAuditError(c, decision)
		return false
	}
	return true
}

func (h *AsyncImageHandler) Get(c *gin.Context) {
	// Polling deliberately does not require the feature to be enabled, only that
	// the task store is reachable. Turning the switch off in the admin UI must not
	// strand tasks that were already accepted — their results are still in Redis
	// and their submitters are still polling.
	if !h.pollable() {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "async image tasks are not enabled")
		return
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.UserID == "" || apiKey.ID == "" {
		imageTaskError(c, service.ErrImageTaskForbidden)
		return
	}
	task, err := h.tasks.Get(c.Request.Context(), service.ImageTaskOwner{UserID: apiKey.UserID, APIKeyID: apiKey.ID}, c.Param("task_id"))
	if err != nil {
		imageTaskError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	if task.Status == service.ImageTaskStatusProcessing {
		c.Header("Retry-After", "3")
	}
	c.JSON(http.StatusOK, task)
}

func (h *AsyncImageHandler) validateRequest(c *gin.Context, platform string, body []byte) error {
	if h.openAI == nil || h.openAI.gatewayService == nil {
		return nil
	}
	if platform == service.PlatformGrok {
		parsed := service.ParseGrokMediaRequest(c.GetHeader("Content-Type"), body)
		if strings.TrimSpace(parsed.Model) == "" {
			return errors.New("model is required")
		}
		return nil
	}
	parsed, err := h.openAI.gatewayService.ParseOpenAIImagesRequest(c, body)
	if err != nil {
		return err
	}
	if parsed.Stream {
		return errors.New("streaming image requests cannot be submitted as asynchronous tasks")
	}
	return nil
}

func (h *AsyncImageHandler) executeWithGateway(platform string, c *gin.Context) {
	if h.openAI == nil {
		imageTaskJSONError(c, http.StatusServiceUnavailable, "api_error", "image gateway is unavailable")
		return
	}
	if platform == service.PlatformGrok {
		h.openAI.GrokImages(c)
		return
	}
	h.openAI.Images(c)
}

func (h *AsyncImageHandler) run(taskID, platform string, taskCtx *gin.Context, recorder *httptest.ResponseRecorder, cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.L().Error("image_task.execution_panicked", zap.String("task_id", taskID), zap.Any("panic", recovered))
			h.failTask(taskID, http.StatusInternalServerError, imageTaskErrorPayload("api_error", "image generation task panicked"))
		}
	}()

	h.execute(platform, taskCtx)
	body := bytes.TrimSpace(recorder.Body.Bytes())
	if err := taskCtx.Request.Context().Err(); err != nil && len(body) == 0 {
		h.failTask(taskID, http.StatusGatewayTimeout, imageTaskErrorPayload("timeout_error", "image generation task timed out"))
		return
	}
	statusCode := recorder.Code
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		if len(body) == 0 || !json.Valid(body) {
			h.failTask(taskID, http.StatusBadGateway, imageTaskErrorPayload("api_error", "upstream returned an invalid image response"))
			return
		}
		if err := h.tasks.Complete(context.Background(), taskID, statusCode, json.RawMessage(body)); err != nil {
			logger.L().Error("image_task.complete_store_failed", zap.String("task_id", taskID), zap.Error(err))
		}
		return
	}
	h.failTask(taskID, statusCode, extractImageTaskError(body))
}

func (h *AsyncImageHandler) failTask(taskID string, statusCode int, taskErr json.RawMessage) {
	if err := h.tasks.Fail(context.Background(), taskID, statusCode, taskErr); err != nil {
		logger.L().Error("image_task.failure_store_failed", zap.String("task_id", taskID), zap.Error(err))
	}
}

func newAsyncImageContext(c *gin.Context, body []byte, timeoutDuration time.Duration) (*gin.Context, *httptest.ResponseRecorder, context.CancelFunc) {
	base := context.WithoutCancel(c.Request.Context())
	executionCtx, cancel := context.WithTimeout(base, timeoutDuration)
	request := c.Request.Clone(executionCtx)
	request.Body = io.NopCloser(bytes.NewReader(body))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	request.ContentLength = int64(len(body))
	request.URL.Path = strings.TrimSuffix(request.URL.Path, "/async")

	taskCtx := c.Copy()
	recorder := httptest.NewRecorder()
	recorderCtx, _ := gin.CreateTestContext(recorder)
	taskCtx.Writer = recorderCtx.Writer
	taskCtx.Request = request
	return taskCtx, recorder, cancel
}

func asyncImageRequestStreams(contentType string, body []byte) bool {
	if isMultipartImagesContentType(contentType) {
		return false
	}
	var envelope struct {
		Stream bool `json:"stream"`
	}
	return json.Unmarshal(body, &envelope) == nil && envelope.Stream
}

func imageTaskPollURL(submitPath, taskID string) string {
	if strings.HasPrefix(submitPath, "/v1/") {
		return "/v1/images/tasks/" + taskID
	}
	return "/images/tasks/" + taskID
}

func extractImageTaskError(body []byte) json.RawMessage {
	if json.Valid(body) {
		var envelope struct {
			Error json.RawMessage `json:"error"`
		}
		if json.Unmarshal(body, &envelope) == nil && len(envelope.Error) > 0 && json.Valid(envelope.Error) {
			return envelope.Error
		}
		return json.RawMessage(body)
	}
	return imageTaskErrorPayload("api_error", "image generation failed")
}

func imageTaskErrorPayload(errorType, message string) json.RawMessage {
	data, _ := json.Marshal(gin.H{"type": errorType, "message": message})
	return data
}

func imageTaskError(c *gin.Context, err error) {
	status := infraerrors.Code(err)
	code := infraerrors.Reason(err)
	message := infraerrors.Message(err)
	if status <= 0 {
		status = http.StatusInternalServerError
	}
	if strings.TrimSpace(code) == "" {
		code = "IMAGE_TASK_ERROR"
	}
	imageTaskJSONError(c, status, code, message)
}

func imageTaskJSONError(c *gin.Context, status int, code, message string) {
	c.Header("Cache-Control", "no-store")
	c.JSON(status, gin.H{"error": gin.H{"type": code, "code": code, "message": message}})
}

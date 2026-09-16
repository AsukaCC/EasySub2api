package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// Probes use the same model routing and protocol adapters as normal inference.
func (s *AccountTestService) testOpenCodeAccountConnection(c *gin.Context, account *Account, model, prompt string) error {
	if err := ValidateOpenCodeAccount(account); err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	if strings.TrimSpace(model) == "" {
		model = DefaultOpenCodeGoTestModel
	}
	if strings.TrimSpace(prompt) == "" {
		prompt = "Reply OK"
	}
	body, err := json.Marshal(map[string]any{"model": model, "input": prompt, "stream": false})
	if err != nil {
		return err
	}
	recorder := httptest.NewRecorder()
	probe, _ := gin.CreateTestContext(recorder)
	probe.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body)).WithContext(c.Request.Context())
	gateway := s.openAIGatewayService
	if gateway == nil {
		gateway = &OpenAIGatewayService{cfg: s.cfg, httpUpstream: s.httpUpstream}
	}
	s.sendEvent(c, TestEvent{Type: "test_start", Model: model})
	_, err = gateway.Forward(probe.Request.Context(), probe, account, body)
	if err != nil || recorder.Code >= 400 {
		return s.sendErrorAndEnd(c, "OpenCode upstream probe failed")
	}
	output := gjson.GetBytes(recorder.Body.Bytes(), "output").Array()
	for _, item := range output {
		for _, content := range item.Get("content").Array() {
			if text := content.Get("text").String(); text != "" {
				s.sendEvent(c, TestEvent{Type: "content", Text: text})
			}
		}
	}
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true})
	return nil
}

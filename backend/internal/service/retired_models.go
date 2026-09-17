package service

import (
	"context"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/openai"
	"github.com/tidwall/gjson"
	"net/http"
)

var ErrModelRetired = infraerrors.New(http.StatusBadRequest, "MODEL_RETIRED", "GPT-5.4 and GPT-5.5 model families have been retired")

func CheckActiveModel(models ...string) error {
	for _, model := range models {
		if openai.IsRetiredModel(model) {
			return ErrModelRetired
		}
	}
	return nil
}

func CheckActiveRequestModel(ctx context.Context, models ...string) error {
	public, _ := RequestedPublicModelFromContext(ctx)
	upstream, _ := ResolvedUpstreamModelFromContext(ctx)
	if err := CheckActiveModel(public, upstream); err != nil {
		return err
	}
	return CheckActiveModel(models...)
}

func CheckActiveAccountModel(account *Account, model string) error {
	if err := CheckActiveModel(model); err != nil {
		return err
	}
	if account != nil && !account.IsOpenAIPassthroughEnabled() {
		return CheckActiveModel(account.GetMappedModel(model))
	}
	return nil
}

func checkActivePayloadModels(account *Account, body []byte) error {
	for _, path := range []string{"model", "session.model", "response.model"} {
		if model := gjson.GetBytes(body, path); model.Type == gjson.String {
			if err := CheckActiveAccountModel(account, model.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIUsageFromGJSONParsesExplicitImageCacheDetails(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantInput      int
		wantImageInput int
		wantCacheRead  int
		wantImageCache int
	}{
		{
			name:           "responses input token details",
			body:           `{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens":50,"cached_tokens_details":{"text_tokens":10,"image_tokens":40}}}`,
			wantInput:      100,
			wantImageInput: 80,
			wantCacheRead:  50,
			wantImageCache: 40,
		},
		{
			name:           "chat prompt token details",
			body:           `{"prompt_tokens":90,"prompt_tokens_details":{"image_tokens":60,"cached_tokens":30,"cached_tokens_details":{"text_tokens":12,"image_tokens":18}}}`,
			wantInput:      90,
			wantImageInput: 60,
			wantCacheRead:  30,
			wantImageCache: 18,
		},
		{
			name:           "details supply compatible cache total",
			body:           `{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens_details":{"text_tokens":10,"image_tokens":40}}}`,
			wantInput:      100,
			wantImageInput: 80,
			wantCacheRead:  50,
			wantImageCache: 40,
		},
		{
			name:           "image cache clamps to cache total",
			body:           `{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens":25,"cached_tokens_details":{"image_tokens":40}}}`,
			wantInput:      100,
			wantImageInput: 80,
			wantCacheRead:  25,
			wantImageCache: 25,
		},
		{
			name:           "image cache clamps to image input",
			body:           `{"input_tokens":100,"input_tokens_details":{"image_tokens":20,"cached_tokens":50,"cached_tokens_details":{"image_tokens":40}}}`,
			wantInput:      100,
			wantImageInput: 20,
			wantCacheRead:  50,
			wantImageCache: 20,
		},
		{
			name:           "cache total does not imply image cache",
			body:           `{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens":50}}`,
			wantInput:      100,
			wantImageInput: 80,
			wantCacheRead:  50,
			wantImageCache: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage, ok := openAIUsageFromGJSON(gjson.Parse(tt.body))
			require.True(t, ok)
			require.Equal(t, tt.wantInput, usage.InputTokens)
			require.Equal(t, tt.wantImageInput, usage.ImageInputTokens)
			require.Equal(t, tt.wantCacheRead, usage.CacheReadInputTokens)
			require.Equal(t, tt.wantImageCache, usage.ImageCacheReadTokens)
		})
	}
}

func TestExtractOpenAIUsageParsesResponsesImageCacheDetails(t *testing.T) {
	usage, ok := extractOpenAIUsageFromJSONBytes([]byte(`{"response":{"usage":{"input_tokens":64,"input_tokens_details":{"image_tokens":48,"cached_tokens":24,"cached_tokens_details":{"text_tokens":8,"image_tokens":16}}}}}`))
	require.True(t, ok)
	require.Equal(t, 64, usage.InputTokens)
	require.Equal(t, 48, usage.ImageInputTokens)
	require.Equal(t, 24, usage.CacheReadInputTokens)
	require.Equal(t, 16, usage.ImageCacheReadTokens)
}

func TestMergeOpenAIUsagePreservesAndDoesNotDoubleImageCacheTokens(t *testing.T) {
	dst := OpenAIUsage{ImageCacheReadTokens: 7}
	withoutImageDetails := []byte(`{"usage":{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens":50}}}`)
	mergeOpenAIUsage(&dst, withoutImageDetails)
	require.Equal(t, 7, dst.ImageCacheReadTokens)

	withImageDetails := []byte(`{"usage":{"input_tokens":100,"input_tokens_details":{"image_tokens":80,"cached_tokens":50,"cached_tokens_details":{"text_tokens":10,"image_tokens":40}}}}`)
	mergeOpenAIUsage(&dst, withImageDetails)
	require.Equal(t, 40, dst.ImageCacheReadTokens)
	mergeOpenAIUsage(&dst, withImageDetails)
	require.Equal(t, 40, dst.ImageCacheReadTokens)
}

func TestCodexDirectImagesUsageUsesCentralImageCacheParser(t *testing.T) {
	usage, ok := codexDirectImagesUsage([]byte(`{"usage":{"input_tokens":100,"input_tokens_details":{"text_tokens":20,"image_tokens":80,"cached_tokens_details":{"text_tokens":10,"image_tokens":40}},"output_tokens":200}}`))
	require.True(t, ok)
	require.Equal(t, 50, usage.CacheReadInputTokens)
	require.Equal(t, 40, usage.ImageCacheReadTokens)
	require.Equal(t, 200, usage.ImageOutputTokens)
}

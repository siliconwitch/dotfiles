// Package claude asks Claude about images.
package claude

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

var ErrRefused = errors.New("Claude declined to answer")

// Ask sends JPEG images and a question, and returns the text of the answer. A schema, when given, makes the
// answer JSON that matches it.
func Ask(key, model, system string, images [][]byte, question string, schema map[string]any) (string, error) {
	client := anthropic.NewClient(option.WithAPIKey(key), option.WithMaxRetries(6))

	var blocks []anthropic.ContentBlockParamUnion
	for _, image := range images {
		blocks = append(blocks, anthropic.NewImageBlockBase64("image/jpeg", base64.StdEncoding.EncodeToString(image)))
	}
	blocks = append(blocks, anthropic.NewTextBlock(question))

	outputConfig := anthropic.OutputConfigParam{Effort: anthropic.OutputConfigEffortLow}
	if schema != nil {
		outputConfig.Format = anthropic.JSONOutputFormatParam{Schema: schema}
	}

	message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:        model,
		MaxTokens:    4096,
		OutputConfig: outputConfig,
		System:       []anthropic.TextBlockParam{{Text: system}},
		Messages:     []anthropic.MessageParam{anthropic.NewUserMessage(blocks...)},
	})
	if err != nil {
		return "", err
	}

	if message.StopReason == anthropic.StopReasonRefusal {
		return "", ErrRefused
	}

	var text []string
	for _, block := range message.Content {
		textBlock, ok := block.AsAny().(anthropic.TextBlock)
		if ok {
			text = append(text, textBlock.Text)
		}
	}

	return strings.Join(text, ""), nil
}

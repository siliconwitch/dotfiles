// Package describe describes every extracted frame with Claude through the Message Batches API,
// which costs half as much as single requests.
package describe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	"video-edit/internal/files"
	"video-edit/internal/secret"
)

func Run(folder string, model string) error {
	_, err := os.Stat(filepath.Join(folder, "descriptions.txt"))
	if err == nil {
		fmt.Println("descriptions.txt exists, skipping")
		return nil
	}

	encodedIndex, err := os.ReadFile(filepath.Join(folder, "frames", "index.json"))
	if err != nil {
		return fmt.Errorf("run frames first: %w", err)
	}

	var index []files.Frame
	err = json.Unmarshal(encodedIndex, &index)
	if err != nil {
		return err
	}

	key, err := secret.Pass("api/anthropic")
	if err != nil {
		return err
	}

	client := anthropic.NewClient(option.WithAPIKey(key))
	ctx := context.Background()

	prompts := map[string]string{
		"screen": "You describe frames from a screen recording for a video editor who cannot see them. " +
			"In one or two sentences, name the application and the file or page shown, and say what is happening, " +
			"such as code being typed, a command running or a result appearing. " +
			"Quote short visible text that identifies the moment. Reply with the description only.",
		"webcam": "You describe frames from a presenter's webcam for a video editor who cannot see them. " +
			"In one sentence, say what the presenter is doing, such as talking to the camera, holding up a board, " +
			"pointing at something or looking at the screen. Call the person \"the presenter\" and never guess their gender. " +
			"Reply with the description only.",
	}

	// The batch ID is saved as soon as it exists, so an interrupted run resumes the paid batch instead of resubmitting it
	batchFile := filepath.Join(folder, "describe-batch.txt")
	batchID, err := os.ReadFile(batchFile)
	if os.IsNotExist(err) {
		var requests []anthropic.MessageBatchNewParamsRequest
		for i, frame := range index {
			image, err := os.ReadFile(filepath.Join(folder, frame.File))
			if err != nil {
				return err
			}

			requests = append(requests, anthropic.MessageBatchNewParamsRequest{
				CustomID: strconv.Itoa(i),
				Params: anthropic.MessageBatchNewParamsRequestParams{
					Model:        model,
					MaxTokens:    2048,
					OutputConfig: anthropic.OutputConfigParam{Effort: anthropic.OutputConfigEffortLow},
					System:       []anthropic.TextBlockParam{{Text: prompts[frame.Stream]}},
					Messages: []anthropic.MessageParam{anthropic.NewUserMessage(
						anthropic.NewImageBlockBase64("image/jpeg", base64.StdEncoding.EncodeToString(image)),
						anthropic.NewTextBlock("Describe this frame."),
					)},
				},
			})
		}

		batch, err := client.Messages.Batches.New(ctx, anthropic.MessageBatchNewParams{Requests: requests})
		if err != nil {
			return err
		}

		batchID = []byte(batch.ID)
		err = os.WriteFile(batchFile, batchID, 0o644)
		if err != nil {
			return err
		}

		fmt.Println("submitted", len(requests), "frames with", model)
	} else if err != nil {
		return err
	}

	for {
		batch, err := client.Messages.Batches.Get(ctx, string(batchID), anthropic.MessageBatchGetParams{})
		if err != nil {
			return err
		}

		if batch.ProcessingStatus == anthropic.MessageBatchProcessingStatusEnded {
			break
		}

		fmt.Println("waiting for Claude:", batch.RequestCounts.Processing, "frames left")
		time.Sleep(30 * time.Second)
	}

	descriptions := make([]string, len(index))
	stream := client.Messages.Batches.ResultsStreaming(ctx, string(batchID), anthropic.MessageBatchResultsParams{})
	for stream.Next() {
		result := stream.Current()

		i, err := strconv.Atoi(result.CustomID)
		if err != nil || i < 0 || i >= len(index) {
			return fmt.Errorf("unexpected result for request %q", result.CustomID)
		}

		if result.Result.Type != "succeeded" {
			descriptions[i] = "(no description: " + result.Result.Type + ")"
			continue
		}

		if result.Result.Message.StopReason == anthropic.StopReasonRefusal {
			descriptions[i] = "(no description: refused)"
			continue
		}

		var text []string
		for _, block := range result.Result.Message.Content {
			textBlock, ok := block.AsAny().(anthropic.TextBlock)
			if ok {
				text = append(text, textBlock.Text)
			}
		}
		descriptions[i] = strings.Join(strings.Fields(strings.Join(text, " ")), " ")
	}

	err = stream.Err()
	if err != nil {
		return err
	}

	var lines strings.Builder
	for i, frame := range index {
		fmt.Fprintf(&lines, "%.1f %s %s: %s\n", frame.Time, frame.Stream, frame.File, descriptions[i])
	}

	err = os.WriteFile(filepath.Join(folder, "descriptions.txt"), []byte(lines.String()), 0o644)
	if err != nil {
		return err
	}

	fmt.Println("wrote", len(index), "descriptions to descriptions.txt")
	return nil
}

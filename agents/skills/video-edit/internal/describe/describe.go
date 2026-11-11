// Package describe describes every extracted frame with Claude. Each description names up to two
// subjects with a box, which render frames the shot around.
package describe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	"video-edit/internal/files"
	"video-edit/internal/project"
)

func Run(folder, model, key string, sources map[string]project.Source) error {
	sorted := project.Sorted(sources)

	indexes := map[string][]files.Frame{}
	for _, source := range sorted {
		encoded, err := os.ReadFile(filepath.Join(folder, ".analysis", source.Key(), "frames.json"))
		if err != nil {
			return fmt.Errorf("run frames first: %w", err)
		}

		var index []files.Frame
		err = json.Unmarshal(encoded, &index)
		if err != nil {
			return err
		}
		indexes[source.Key()] = index
	}

	subjects := "Then list the subjects. Give each a short label and a box: its left, top, right and bottom " +
		"edges in thousandths of the image width and height, so [0, 0, 1000, 1000] is the whole image. Draw the box " +
		"tight around the subject. For a person, the box covers the head and shoulders. List at most two subjects. "
	prompts := map[string]string{
		"screen": "You describe frames from a screen recording for a video editor who cannot see them. " +
			"In one or two sentences, name the application and the file or page shown, and say what is happening, " +
			"such as code being typed, a command running or a result appearing. " +
			"Quote short visible text that identifies the moment. " + subjects +
			"A subject here is a small region the viewer must read to follow the moment, such as a dialog, a settings " +
			"window, a popup or a few lines of output that fill a small part of the screen. When the whole screen " +
			"matters, such as an editor full of code, list no subjects.",
		"webcam": "You describe frames from a presenter's webcam for a video editor who cannot see them. " +
			"In one sentence, say what the presenter is doing, such as talking to the camera, holding up a board, " +
			"pointing at something or looking at the screen. Call the person \"the presenter\" and never guess their gender. " +
			subjects + "The subjects are the presenter, and any object the presenter shows to the camera, such as a board. " +
			"List no subject when the presenter is out of frame.",
		"video": "You describe frames from b-roll, such as a phone clip of an event, for a video editor who cannot see them. " +
			"In one sentence, say where the shot is and what is happening. Quote short visible text, such as a sign or " +
			"a slide title, that names the place or event. Never guess anyone's name or gender. " + subjects +
			"The subjects are the people or objects the shot is about, such as a speaker on stage or a board on a table, each boxed whole. " +
			"List no subject for a wide shot of a scene. When the shot is of a screen, list a subject only for a small region " +
			"the viewer must read, such as a dialog, and none when the whole screen matters.",
		"photo": "You describe photos for a video editor who cannot see them. In one sentence, say what the photo shows. " +
			"Quote short visible text, such as a part number or a sign. Never guess anyone's name or gender. " + subjects +
			"The subjects are the people or objects the photo is about, each boxed whole, such as a whole device rather than one part of it.",
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"description": map[string]any{"type": "string"},
			"subjects": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"label": map[string]any{"type": "string"},
						"box":   map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
					},
					"required":             []string{"label", "box"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"description", "subjects"},
		"additionalProperties": false,
	}

	type job struct {
		key   string
		index int
	}
	var jobs []job
	for _, source := range sorted {
		for i, frame := range indexes[source.Key()] {
			if frame.Description == "" {
				jobs = append(jobs, job{source.Key(), i})
			}
		}
	}

	if len(jobs) > 0 {
		client := anthropic.NewClient(option.WithAPIKey(key))
		fmt.Println("describing", len(jobs), "frames with", model)

		// Eight requests at a time. Each frame belongs to one worker, so the workers share only the queue.
		queue := make(chan job)
		failures := make(chan error, len(jobs))
		var workers sync.WaitGroup
		for range 8 {
			workers.Go(func() {
				for job := range queue {
					frame := &indexes[job.key][job.index]
					image, err := os.ReadFile(filepath.Join(folder, frame.File))
					if err != nil {
						failures <- err
						continue
					}

					message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
						Model:     model,
						MaxTokens: 2048,
						OutputConfig: anthropic.OutputConfigParam{
							Effort: anthropic.OutputConfigEffortLow,
							Format: anthropic.JSONOutputFormatParam{Schema: schema},
						},
						System: []anthropic.TextBlockParam{{Text: prompts[frame.Stream]}},
						Messages: []anthropic.MessageParam{anthropic.NewUserMessage(
							anthropic.NewImageBlockBase64("image/jpeg", base64.StdEncoding.EncodeToString(image)),
							anthropic.NewTextBlock("Describe this frame."),
						)},
					})
					if err != nil {
						failures <- fmt.Errorf("%s: %w", frame.File, err)
						continue
					}

					if message.StopReason == anthropic.StopReasonRefusal {
						frame.Description = "(no description: refused)"
						continue
					}

					var text []string
					for _, block := range message.Content {
						textBlock, ok := block.AsAny().(anthropic.TextBlock)
						if ok {
							text = append(text, textBlock.Text)
						}
					}

					var described struct {
						Description string `json:"description"`
						Subjects    []struct {
							Label string    `json:"label"`
							Box   []float64 `json:"box"`
						} `json:"subjects"`
					}
					err = json.Unmarshal([]byte(strings.Join(text, "")), &described)
					if err != nil {
						frame.Description = "(no description: unreadable reply)"
						continue
					}

					frame.Description = strings.Join(strings.Fields(described.Description), " ")
					for _, subject := range described.Subjects {
						if len(subject.Box) != 4 || len(frame.Subjects) == 2 {
							continue
						}

						var box [4]float64
						for j, edge := range subject.Box {
							box[j] = math.Max(0, math.Min(1000, edge))
						}
						if box[0] >= box[2] || box[1] >= box[3] {
							continue
						}

						frame.Subjects = append(frame.Subjects, files.Subject{Label: subject.Label, Box: box})
					}
				}
			})
		}
		for _, job := range jobs {
			queue <- job
		}
		close(queue)
		workers.Wait()
		close(failures)

		// The frames described so far are saved even when some failed, so a second run asks only for the rest
		for key, index := range indexes {
			encoded, err := json.MarshalIndent(index, "", "  ")
			if err != nil {
				return err
			}

			err = os.WriteFile(filepath.Join(folder, ".analysis", key, "frames.json"), encoded, 0o644)
			if err != nil {
				return err
			}
		}

		failed := len(failures)
		if failed > 0 {
			return fmt.Errorf("%d of %d frames failed, such as %w. Run analyse again for the rest", failed, len(jobs), <-failures)
		}
	}

	// descriptions.txt is for reading: each source's details, then every frame with its subjects' labels
	var lines strings.Builder
	for _, source := range sorted {
		details := []string{source.Kind, fmt.Sprintf("%dx%d", source.Width, source.Height)}
		if source.Duration > 0 {
			details = append(details, fmt.Sprintf("%.1f s", source.Duration))
		}
		if source.Audio {
			details = append(details, "sound")
		}
		if source.Created != "" {
			details = append(details, "created "+source.Created)
		}
		if source.Location != "" {
			details = append(details, "at "+source.Location)
		}
		fmt.Fprintf(&lines, "== %s: %s ==\n", source.Path, strings.Join(details, ", "))
		for _, frame := range indexes[source.Key()] {
			var labels []string
			for _, subject := range frame.Subjects {
				labels = append(labels, subject.Label)
			}
			fmt.Fprintf(&lines, "%.1f %s %s: %s [%s]\n", frame.Time, frame.Stream, frame.File, frame.Description, strings.Join(labels, "; "))
		}
	}

	err := os.WriteFile(filepath.Join(folder, ".analysis", "descriptions.txt"), []byte(lines.String()), 0o644)
	if err != nil {
		return err
	}

	fmt.Println("wrote .analysis/descriptions.txt")
	return nil
}

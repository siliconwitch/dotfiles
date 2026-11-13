// Package preanalyse writes a short summary of each candidate file, so a sub-agent can shortlist the footage
// worth analysing in full.
package preanalyse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"video-edit/internal/claude"
	"video-edit/internal/media"
)

// Run summarises each file from four frames spread across it, or from the image itself, with Haiku, and writes
// the summaries in the order given.
func Run(output string, files []string, key string) error {
	temporary, err := os.MkdirTemp("", "pre-analyse-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	summaries := make([]string, len(files))
	queue := make(chan int)
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			for i := range queue {
				file := files[i]
				info, err := media.Probe(file)
				if err != nil {
					summaries[i] = fmt.Sprintf("== %s\nnot footage: %v\n", file, err)
					continue
				}

				details := []string{"video", fmt.Sprintf("%dx%d", info.Width, info.Height)}
				if info.Image {
					details[0] = "image"
				} else {
					details = append(details, fmt.Sprintf("%d:%02d", int(info.Duration)/60, int(info.Duration)%60))
				}
				if info.Audio {
					details = append(details, "sound")
				}
				details = append(details, "recorded "+info.Recorded.In(time.Local).Format("2006-01-02 15:04 -07:00")+" from "+info.RecordedFrom)
				if info.Location != "" {
					details = append(details, "at "+info.Location)
				}

				var images [][]byte
				positions := []float64{0.125, 0.375, 0.625, 0.875}
				if info.Image {
					positions = []float64{0}
				}
				for j, position := range positions {
					frame := filepath.Join(temporary, fmt.Sprintf("%d-%d.jpg", i, j))
					args := []string{"-ss", fmt.Sprintf("%.3f", position*info.Duration), "-i", file}
					if info.Image {
						args = []string{"-i", file}
					}
					_, err = media.Run(append(args, "-frames:v", "1", "-vf", "scale=768:768:force_original_aspect_ratio=decrease", "-q:v", "4", frame)...)
					if err != nil {
						break
					}

					var image []byte
					image, err = os.ReadFile(frame)
					if err != nil {
						break
					}
					images = append(images, image)
				}
				if err != nil {
					summaries[i] = fmt.Sprintf("== %s\n%s\nno frames: %v\n", file, strings.Join(details, ", "), err)
					continue
				}

				summary, err := claude.Ask(key, "claude-haiku-5-5",
					"You summarise footage for a video editor choosing which files to use. In one or two sentences, say "+
						"what the footage shows: the people, place, devices and screens, and any visible text that names the "+
						"subject. Call each person \"a person\". Never guess anyone's name or gender.",
					images, fmt.Sprintf("These are %d frames spread across one file. Summarise it.", len(images)), nil)
				if err != nil {
					summary = fmt.Sprintf("no summary: %v", err)
				}

				summaries[i] = fmt.Sprintf("== %s\n%s\n%s\n", file, strings.Join(details, ", "), strings.Join(strings.Fields(summary), " "))
			}
		})
	}
	for i := range files {
		queue <- i
	}
	close(queue)
	workers.Wait()

	err = os.MkdirAll(filepath.Dir(output), 0o755)
	if err != nil {
		return err
	}

	err = os.WriteFile(output, []byte(strings.Join(summaries, "\n")), 0o644)
	if err != nil {
		return err
	}

	fmt.Printf("wrote %d summaries to %s\n", len(files), output)
	return nil
}

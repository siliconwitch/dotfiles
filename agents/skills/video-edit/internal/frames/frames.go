// Package frames extracts the frames that describe sends to Claude, from every stream of every source.
package frames

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"video-edit/internal/ffmpeg"
	"video-edit/internal/files"
	"video-edit/internal/project"
)

func Run(folder string, sources map[string]project.Source) error {
	showinfoTime := regexp.MustCompile(`\[showinfo@(\w+) @ [^\]]+\] n:.*pts_time:\s*([0-9.]+)`)

	for _, source := range project.Sorted(sources) {
		directory := filepath.Join(".analysis", source.Key())
		_, err := os.Stat(filepath.Join(folder, directory, "frames.json"))
		if err == nil {
			continue
		}

		err = os.MkdirAll(filepath.Join(folder, directory, "frames"), 0o755)
		if err != nil {
			return err
		}

		// One decode feeds every stream, because decoding a recording costs far more than the filters.
		// The screen changes in steps, so take a frame when it changes, at most one a second and at least
		// one every 10 s. A camera changes continuously, so take one every 2 s, often enough for framing to
		// see a board held up. showinfo logs each kept frame's time. 1568 px is the longest edge Claude
		// reads unscaled, and 1280 px is enough to place a subject.
		streams := source.Streams()
		graph := fmt.Sprintf("[0:v:0]split=%d", len(streams))
		for i := range streams {
			graph += fmt.Sprintf("[in%d]", i)
		}
		var outputs []string
		for i, stream := range streams {
			selection, longEdge := "fps=1/2", 1280
			switch stream.Name {
			case "screen":
				selection, longEdge = "select='isnan(prev_selected_t)+gte(t-prev_selected_t,10)+gt(scene,0.1)*gte(t-prev_selected_t,1)'", 1568
			case "photo":
				selection, longEdge = "select='eq(n,0)'", 1568
			}

			scale := fmt.Sprintf("scale=%d:-2", longEdge)
			if stream.Height > stream.Width {
				scale = fmt.Sprintf("scale=-2:%d", longEdge)
			}

			graph += fmt.Sprintf(";[in%d]%s,%s,%s,showinfo@%s[out%d]", i, source.Filters(stream), scale, selection, stream.Name, i)
			outputs = append(outputs, "-map", fmt.Sprintf("[out%d]", i), "-fps_mode", "vfr", "-q:v", "4", filepath.Join(directory, "frames", stream.Name+"-%06d.jpg"))
		}

		fmt.Println("extracting frames from", source.Path)
		log, err := ffmpeg.Run(folder, append([]string{"-i", filepath.Join(".sources", source.Path), "-filter_complex", graph}, outputs...)...)
		if err != nil {
			return err
		}

		var index []files.Frame
		counts := map[string]int{}
		for _, match := range showinfoTime.FindAllStringSubmatch(log, -1) {
			time, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				return err
			}

			counts[match[1]]++
			file := filepath.Join(directory, "frames", fmt.Sprintf("%s-%06d.jpg", match[1], counts[match[1]]))
			index = append(index, files.Frame{Stream: match[1], Time: time, File: file})
		}

		sort.SliceStable(index, func(i, j int) bool { return index[i].Time < index[j].Time })

		encoded, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(folder, directory, "frames.json"), encoded, 0o644)
		if err != nil {
			return err
		}

		fmt.Println("wrote", len(index), "frames to", filepath.Join(directory, "frames.json"))
	}

	return nil
}

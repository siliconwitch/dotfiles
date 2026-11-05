// Package frames extracts the frames that describe sends to Claude.
package frames

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"video-edit/internal/canvas"
	"video-edit/internal/ffmpeg"
	"video-edit/internal/files"
)

func Run(folder string) error {
	_, err := os.Stat(filepath.Join(folder, "frames", "index.json"))
	if err == nil {
		fmt.Println("frames/index.json exists, skipping")
		return nil
	}

	err = os.MkdirAll(filepath.Join(folder, "frames"), 0o755)
	if err != nil {
		return err
	}

	// The screen changes in steps, so take a frame when it changes, at most one a second and at least
	// one every 10 s. The webcam changes continuously, so take one every 5 s. showinfo logs each
	// kept frame's time, and the long edge of 1568 px is the largest image Claude reads unscaled.
	streams := []struct {
		name    string
		filters string
	}{
		{"screen", "crop=" + canvas.Screen + ",select='isnan(prev_selected_t)+gte(t-prev_selected_t,10)+gt(scene,0.1)*gte(t-prev_selected_t,1)',showinfo,scale=1568:-2"},
		{"webcam", "crop=" + canvas.Webcam + ",fps=1/5,showinfo,scale=640:-2"},
	}

	showinfoTime := regexp.MustCompile(`Parsed_showinfo.*pts_time:\s*([0-9.]+)`)

	var index []files.Frame
	for _, stream := range streams {
		fmt.Println("extracting", stream.name, "frames")
		log, err := ffmpeg.Run(folder, "-i", "source.mkv", "-map", "0:v:0", "-vf", stream.filters, "-fps_mode", "vfr", "-q:v", "4", "frames/"+stream.name+"-%06d.jpg")
		if err != nil {
			return err
		}

		for i, match := range showinfoTime.FindAllStringSubmatch(log, -1) {
			time, err := strconv.ParseFloat(match[1], 64)
			if err != nil {
				return err
			}

			index = append(index, files.Frame{Stream: stream.name, Time: time, File: fmt.Sprintf("frames/%s-%06d.jpg", stream.name, i+1)})
		}
	}

	sort.SliceStable(index, func(i, j int) bool { return index[i].Time < index[j].Time })

	encoded, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(folder, "frames", "index.json"), encoded, 0o644)
	if err != nil {
		return err
	}

	fmt.Println("wrote", len(index), "frames to frames/index.json")
	return nil
}

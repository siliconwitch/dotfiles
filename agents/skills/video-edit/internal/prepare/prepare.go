// Package prepare splits each OBS recording in a folder into its screen and webcam footage.
package prepare

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"video-edit/internal/media"
)

// Run splits every recording in the folder not yet split. The OBS canvas is 4480x1440: the screen fills
// 2560x1440 at the left, and the webcam fills 1920x1080 at the top right. Each split file keeps the recording's
// sound and start time. The recording is then renamed with a leading dot, so the next run skips it.
func Run(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}

	split := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasPrefix(name, ".") || strings.HasSuffix(name, "-capture.mkv") {
			continue
		}

		recording := filepath.Join(folder, name)
		started, err := time.ParseInLocation("2006-01-02 15-04-05.mkv", name, time.Local)
		if err != nil {
			fmt.Printf("skipped %s: not named like an OBS recording\n", name)
			continue
		}

		// OBS writes to the file it is recording, so a file changed in the last 30 s is still being recorded
		file, err := entry.Info()
		if err != nil {
			return err
		}
		if time.Since(file.ModTime()) < 30*time.Second {
			fmt.Printf("skipped %s: still recording\n", name)
			continue
		}

		info, err := media.Probe(recording)
		if err != nil {
			return err
		}
		if info.Width != 4480 || info.Height != 1440 {
			fmt.Printf("skipped %s: %dx%d is not the 4480x1440 OBS canvas\n", name, info.Width, info.Height)
			continue
		}

		stem := started.Format("20060102-150405")
		outputs := map[string]string{
			"screen": filepath.Join(folder, stem+"-screen-capture.mkv"),
			"webcam": filepath.Join(folder, stem+"-webcam-capture.mkv"),
		}
		for _, output := range outputs {
			_, err := os.Stat(output)
			if err == nil {
				return fmt.Errorf("%s already exists. Move it away, then run prepare again", output)
			}
		}

		partial := func(output string) string {
			return filepath.Join(folder, "."+strings.TrimSuffix(filepath.Base(output), ".mkv")+".partial.mkv")
		}

		// The GPU encodes both pictures from one decode. Quality 60 keeps each close to the recording's own
		// bitrate, so text on the screen stays sharp.
		creationTime := started.UTC().Format("2006-01-02T15:04:05.000000Z")
		fmt.Printf("splitting %s\n", name)
		_, err = media.Run("-init_hw_device", "vaapi=gpu", "-filter_hw_device", "gpu", "-i", recording,
			"-filter_complex", "[0:v]split[s][w];[s]crop=2560:1440:0:0,format=nv12,hwupload[screen];[w]crop=1920:1080:2560:0,format=nv12,hwupload[webcam]",
			"-map", "[screen]", "-map", "0:a?", "-c:v", "av1_vaapi", "-rc_mode", "CQP", "-global_quality", "60", "-c:a", "copy",
			"-metadata", "creation_time="+creationTime, partial(outputs["screen"]),
			"-map", "[webcam]", "-map", "0:a?", "-c:v", "av1_vaapi", "-rc_mode", "CQP", "-global_quality", "60", "-c:a", "copy",
			"-metadata", "creation_time="+creationTime, partial(outputs["webcam"]))
		if err != nil {
			return fmt.Errorf("splitting %s: %w", name, err)
		}

		for _, output := range outputs {
			err = os.Rename(partial(output), output)
			if err != nil {
				return err
			}
		}

		err = os.Rename(recording, filepath.Join(folder, "."+name))
		if err != nil {
			return err
		}

		fmt.Printf("wrote %s\nwrote %s\n", outputs["screen"], outputs["webcam"])
		split++
	}

	fmt.Printf("split %d recordings in %s\n", split, folder)
	return nil
}

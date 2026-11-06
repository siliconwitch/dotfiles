// Package sheet makes contact sheets of a rendered video: one labelled frame from the middle of every
// segment and cutaway, and one from every title, so the agent sees the whole edit in a few images.
package sheet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"video-edit/internal/ffmpeg"
	"video-edit/internal/project"
)

func Run(folder, format string) error {
	sources, err := project.ProbeAll(folder)
	if err != nil {
		return err
	}

	list, err := project.Load(folder, format, sources)
	if err != nil {
		return err
	}

	video := filepath.Join("out", format+".mp4")
	_, err = os.Stat(filepath.Join(folder, video))
	if err != nil {
		return fmt.Errorf("run render %s first: %w", format, err)
	}

	type moment struct {
		time  float64
		label string
	}
	var moments []moment
	for i, segment := range list.Segments {
		length := segment.End - segment.Start
		moments = append(moments, moment{segment.OutStart + length/2, fmt.Sprintf("segment %d %s %s", i+1, segment.Source, segment.Layout)})
		if segment.Title != "" {
			moments = append(moments, moment{segment.OutStart + 1.5, fmt.Sprintf("segment %d title", i+1)})
		}
	}
	for i, cutaway := range list.Cutaways {
		moments = append(moments, moment{(cutaway.OutStart + cutaway.OutEnd) / 2, fmt.Sprintf("cutaway %d %s", i+1, cutaway.Source)})
		if cutaway.Title != "" {
			moments = append(moments, moment{cutaway.OutStart + 1.5, fmt.Sprintf("cutaway %d title", i+1)})
		}
	}
	sort.Slice(moments, func(a, b int) bool { return moments[a].time < moments[b].time })
	for i := range moments {
		moments[i].time = min(moments[i].time, list.Duration-0.1)
	}

	directory := filepath.Join("build", format, "sheet")
	err = os.RemoveAll(filepath.Join(folder, directory))
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(folder, directory), 0o755)
	if err != nil {
		return err
	}

	tileWidth, columns, rows := 800, 3, 4
	if format == "portrait" {
		tileWidth, columns, rows = 360, 6, 2
	}

	for i, moment := range moments {
		// drawtext reads the label from a file, which needs no escaping
		labelFile := filepath.Join(directory, fmt.Sprintf("%03d.txt", i+1))
		err = os.WriteFile(filepath.Join(folder, labelFile), []byte(fmt.Sprintf("%.1f s  %s", moment.time, moment.label)), 0o644)
		if err != nil {
			return err
		}

		_, err = ffmpeg.Run(folder, "-ss", fmt.Sprintf("%.3f", moment.time), "-i", video, "-frames:v", "1",
			"-vf", fmt.Sprintf("scale=%d:-2:out_range=full,drawtext=font=Geist:textfile=%s:x=10:y=10:fontsize=%d:fontcolor=white:box=1:boxcolor=black@0.7:boxborderw=6", tileWidth, labelFile, tileWidth/36),
			filepath.Join(directory, fmt.Sprintf("%03d.jpg", i+1)))
		if err != nil {
			return err
		}
	}

	_, err = ffmpeg.Run(folder, "-framerate", "1", "-i", filepath.Join(directory, "%03d.jpg"),
		"-vf", fmt.Sprintf("tile=%dx%d:padding=8:margin=8:color=0x100c14", columns, rows), "-q:v", "3",
		filepath.Join("build", format, "sheet-%d.jpg"))
	if err != nil {
		return err
	}

	sheets, err := filepath.Glob(filepath.Join(folder, "build", format, "sheet-*.jpg"))
	if err != nil {
		return err
	}

	for _, sheet := range sheets {
		fmt.Println(sheet)
	}
	return nil
}

// Package media runs ffmpeg and reads what a footage file holds.
package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Info describes a footage file. Width and height are as displayed, after the file's rotation.
type Info struct {
	Image        bool
	Width        int
	Height       int
	Duration     float64
	FPS          float64
	Audio        bool
	HDR          bool
	Recorded     time.Time
	RecordedFrom string
	Location     string
}

// Run runs ffmpeg at the lowest priority, so a long encode leaves the desktop responsive, and returns its log.
func Run(args ...string) (string, error) {
	command := exec.Command("nice", append([]string{"-n", "19", "ffmpeg", "-hide_banner", "-nostdin", "-y"}, args...)...)

	var log bytes.Buffer
	command.Stderr = &log

	err := command.Run()
	if err != nil {
		lines := strings.Split(strings.TrimSpace(log.String()), "\n")
		return "", fmt.Errorf("ffmpeg failed: %w\n%s", err, strings.Join(lines[max(0, len(lines)-20):], "\n"))
	}

	return log.String(), nil
}

// Probe reads the file's picture, sound, length, and when and where it was recorded.
func Probe(path string) (Info, error) {
	output, err := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", "-show_format", path).Output()
	if err != nil {
		return Info{}, fmt.Errorf("ffprobe cannot read %s: %w", path, err)
	}

	var probe struct {
		Streams []struct {
			CodecType     string `json:"codec_type"`
			Width         int    `json:"width"`
			Height        int    `json:"height"`
			AverageRate   string `json:"avg_frame_rate"`
			ColorTransfer string `json:"color_transfer"`
			Disposition   struct {
				AttachedPicture int `json:"attached_pic"`
			} `json:"disposition"`
			SideDataList []struct {
				Rotation float64 `json:"rotation"`
			} `json:"side_data_list"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	err = json.Unmarshal(output, &probe)
	if err != nil {
		return Info{}, err
	}

	var info Info
	pictureFound := false
	for _, stream := range probe.Streams {
		if stream.CodecType == "audio" {
			info.Audio = true
		}
		if stream.CodecType != "video" || stream.Disposition.AttachedPicture == 1 || pictureFound {
			continue
		}

		pictureFound = true
		info.Width, info.Height = stream.Width, stream.Height
		for _, sideData := range stream.SideDataList {
			if math.Abs(math.Mod(sideData.Rotation, 180)) == 90 {
				info.Width, info.Height = stream.Height, stream.Width
			}
		}
		info.HDR = stream.ColorTransfer == "smpte2084" || stream.ColorTransfer == "arib-std-b67"

		numerator, denominator, _ := strings.Cut(stream.AverageRate, "/")
		frames, _ := strconv.ParseFloat(numerator, 64)
		seconds, _ := strconv.ParseFloat(denominator, 64)
		if frames > 0 && seconds > 0 {
			info.FPS = frames / seconds
		}
	}

	if !pictureFound {
		return Info{}, fmt.Errorf("%s has no picture", path)
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".heic", ".heif", ".avif", ".tif", ".tiff":
		info.Image = true
		info.Audio = false
		info.FPS = 0
	default:
		info.Duration, err = strconv.ParseFloat(probe.Format.Duration, 64)
		if err != nil {
			return Info{}, fmt.Errorf("%s has no length", path)
		}
	}

	// exiftool reads the recording time and place from every container, including a photo's EXIF
	output, err = exec.Command("exiftool", "-j", "-api", "QuickTimeUTC", "-d", "%Y-%m-%dT%H:%M:%S%z",
		"-CreationDate", "-DateTimeOriginal", "-CreateDate", "-GPSLatitude#", "-GPSLongitude#", path).Output()
	if err != nil {
		return Info{}, fmt.Errorf("exiftool cannot read %s: %w", path, err)
	}

	var tags []struct {
		CreationDate     string   `json:"CreationDate"`
		DateTimeOriginal string   `json:"DateTimeOriginal"`
		CreateDate       string   `json:"CreateDate"`
		Latitude         *float64 `json:"GPSLatitude"`
		Longitude        *float64 `json:"GPSLongitude"`
	}
	err = json.Unmarshal(output, &tags)
	if err != nil || len(tags) != 1 {
		return Info{}, fmt.Errorf("exiftool gave no tags for %s", path)
	}

	for _, date := range []string{tags[0].CreationDate, tags[0].DateTimeOriginal, tags[0].CreateDate} {
		recorded, err := time.Parse("2006-01-02T15:04:05-0700", date)
		if err == nil && recorded.Year() > 1970 {
			info.Recorded, info.RecordedFrom = recorded, "the file's metadata"
			break
		}
	}

	if info.Recorded.IsZero() {
		file, err := os.Stat(path)
		if err != nil {
			return Info{}, err
		}
		info.Recorded = file.ModTime().Add(-time.Duration(info.Duration * float64(time.Second)))
		info.RecordedFrom = "the file's modified time, roughly"
	}

	if tags[0].Latitude != nil && tags[0].Longitude != nil {
		info.Location = fmt.Sprintf("%.5f,%.5f", *tags[0].Latitude, *tags[0].Longitude)
	}

	return info, nil
}

// Font returns the file of the installed font that best matches a fontconfig pattern, such as
// "DINish Expanded:style=Medium".
func Font(pattern string) (string, error) {
	output, err := exec.Command("fc-match", "-f", "%{file}", pattern).Output()
	if err != nil {
		return "", fmt.Errorf("fc-match cannot find %s: %w", pattern, err)
	}

	return string(output), nil
}

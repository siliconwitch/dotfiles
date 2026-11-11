// Package project reads a project folder: the sources it holds and the edit lists that cut them.
package project

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	Recording = "recording"
	Broll     = "broll"
	Photo     = "photo"
)

// Source is one file in .sources/. Width and height are as displayed, after the file's rotation.
type Source struct {
	Path     string  `json:"path"` // relative to .sources/
	Kind     string  `json:"kind"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Duration float64 `json:"duration,omitempty"`
	Audio    bool    `json:"audio,omitempty"`
	HDR      bool    `json:"hdr,omitempty"`
	Created  string  `json:"created,omitempty"`
	Location string  `json:"location,omitempty"` // ISO 6709 coordinates, as phones record them
}

// Stream is one picture in a source. A recording holds two, cropped from its OBS canvas.
type Stream struct {
	Name   string // "screen", "webcam", "video" or "photo"
	Crop   string // ffmpeg crop arguments, empty for the whole frame
	Width  int
	Height int
}

// Probe describes the file at path. A file the size of the OBS canvas in SKILL.md is a recording.
func Probe(path string) (Source, error) {
	output, err := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", "-show_format", path).Output()
	if err != nil {
		return Source{}, fmt.Errorf("ffprobe cannot read %s: %w", path, err)
	}

	var probe struct {
		Streams []struct {
			CodecType     string `json:"codec_type"`
			Width         int    `json:"width"`
			Height        int    `json:"height"`
			ColorTransfer string `json:"color_transfer"`
			Disposition   struct {
				AttachedPic int `json:"attached_pic"`
			} `json:"disposition"`
			SideDataList []struct {
				Rotation float64 `json:"rotation"`
			} `json:"side_data_list"`
		} `json:"streams"`
		Format struct {
			Duration string            `json:"duration"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
	}
	err = json.Unmarshal(output, &probe)
	if err != nil {
		return Source{}, err
	}

	source := Source{
		Created:  probe.Format.Tags["com.apple.quicktime.creationdate"],
		Location: probe.Format.Tags["com.apple.quicktime.location.ISO6709"],
	}
	if source.Created == "" {
		source.Created = probe.Format.Tags["creation_time"]
	}
	if source.Location == "" {
		source.Location = probe.Format.Tags["location"]
	}

	found := false
	for _, stream := range probe.Streams {
		if stream.CodecType == "audio" {
			source.Audio = true
		}
		if stream.CodecType != "video" || stream.Disposition.AttachedPic == 1 || found {
			continue
		}

		found = true
		source.Width, source.Height = stream.Width, stream.Height
		for _, sideData := range stream.SideDataList {
			if math.Abs(math.Mod(sideData.Rotation, 180)) == 90 {
				source.Width, source.Height = stream.Height, stream.Width
			}
		}
		source.HDR = stream.ColorTransfer == "smpte2084" || stream.ColorTransfer == "arib-std-b67"
	}

	if !found {
		return Source{}, fmt.Errorf("%s has no picture", path)
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".heic", ".heif", ".avif", ".tif", ".tiff":
		source.Kind = Photo
		source.Audio = false
		return source, nil
	}

	source.Duration, err = strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		return Source{}, fmt.Errorf("%s has no duration", path)
	}

	source.Kind = Broll
	if source.Width == 4480 && source.Height == 1440 {
		source.Kind = Recording
	}

	return source, nil
}

// ProbeAll describes every file in the folder's .sources/, keyed by its path relative to .sources/.
func ProbeAll(folder string) (map[string]Source, error) {
	sources := map[string]Source{}
	root := filepath.Join(folder, ".sources")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			return err
		}

		source, err := Probe(path)
		if err != nil {
			return err
		}

		source.Path, err = filepath.Rel(root, path)
		if err != nil {
			return err
		}

		sources[source.Path] = source
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("%s holds no sources. Run ingest first", root)
	}

	return sources, nil
}

// Sorted returns the sources in path order, so every step lists them the same way.
func Sorted(sources map[string]Source) []Source {
	var sorted []Source
	for _, source := range sources {
		sorted = append(sorted, source)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	return sorted
}

// Key names the source's folder in .analysis/, such as broll-IMG_1234 for broll/IMG_1234.MOV.
func (s Source) Key() string {
	return strings.ReplaceAll(strings.TrimSuffix(s.Path, filepath.Ext(s.Path)), "/", "-")
}

// Streams lists the pictures in the source. The recording crops match the OBS canvas in SKILL.md.
func (s Source) Streams() []Stream {
	switch s.Kind {
	case Recording:
		return []Stream{
			{Name: "screen", Crop: "2560:1440:0:0", Width: 2560, Height: 1440},
			{Name: "webcam", Crop: "1920:1080:2560:0", Width: 1920, Height: 1080},
		}
	case Photo:
		return []Stream{{Name: "photo", Width: s.Width, Height: s.Height}}
	default:
		return []Stream{{Name: "video", Width: s.Width, Height: s.Height}}
	}
}

// Stream returns the named stream of the source.
func (s Source) Stream(name string) (Stream, bool) {
	for _, stream := range s.Streams() {
		if stream.Name == name {
			return stream, true
		}
	}
	return Stream{}, false
}

// Filters returns the filters that turn a decoded frame of the source into the stream's SDR picture.
// ffmpeg has already applied the file's rotation.
func (s Source) Filters(stream Stream) string {
	var filters []string
	if s.HDR {
		filters = append(filters, "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv")
	}
	if stream.Crop != "" {
		filters = append(filters, "crop="+stream.Crop)
	}
	return strings.Join(append(filters, "format=yuv420p", "setsar=1"), ",")
}

// Package animate renders stills of an animation just after each cue, in every format that uses it, so
// the animator checks the drawing without waiting for a render.
package animate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"video-edit/internal/chromium"
	"video-edit/internal/project"
)

func Run(folder, id string) error {
	sources, err := project.ProbeAll(folder)
	if err != nil {
		return err
	}

	page, err := filepath.Abs(filepath.Join(folder, "animations", id, "index.html"))
	if err != nil {
		return err
	}

	_, err = os.Stat(page)
	if err != nil {
		return fmt.Errorf("animations/%s has no index.html", id)
	}

	stills := filepath.Join(folder, "build", "animations", id)
	err = os.MkdirAll(stills, 0o755)
	if err != nil {
		return err
	}

	used := false
	for _, format := range []string{"landscape", "portrait"} {
		_, err := os.Stat(filepath.Join(folder, "edit", format+".json"))
		if os.IsNotExist(err) {
			continue
		}

		// Another agent may be writing the other edit list, so a broken one costs only its own stills
		list, err := project.Load(folder, format, sources)
		if err != nil {
			fmt.Printf("no %s stills: %v\n", format, err)
			continue
		}

		output := project.Formats[format]
		for i, cutaway := range list.Cutaways {
			if cutaway.Source != "animations/"+id {
				continue
			}
			used = true

			// At one device pixel per CSS pixel, which shows the layout at half the landscape size
			panel := output.CutawayPanel(cutaway.Layout)
			browser, err := chromium.Open("file://"+page, int(float64(panel.Width)/output.Scale), int(float64(panel.Height)/output.Scale), 1, cutaway.Properties())
			if err != nil {
				return err
			}

			type moment struct {
				name string
				time float64
			}
			moments := []moment{{"start", 0.5}}
			for cue, offset := range cutaway.CueOffsets {
				moments = append(moments, moment{"cue-" + cue, offset + 0.8})
			}
			moments = append(moments, moment{"end", cutaway.OutEnd - cutaway.OutStart - 0.1})
			sort.Slice(moments, func(a, b int) bool { return moments[a].time < moments[b].time })

			for k, moment := range moments {
				picture, err := browser.Frame(moment.time)
				if err != nil {
					browser.Close()
					return err
				}

				file := filepath.Join(stills, fmt.Sprintf("%s-cutaway%d-%02d-%s.png", format, i+1, k+1, moment.name))
				err = os.WriteFile(file, picture, 0o644)
				if err != nil {
					browser.Close()
					return err
				}

				fmt.Printf("%.2f s: %s\n", moment.time, file)
			}
			browser.Close()
		}
	}

	if !used {
		return fmt.Errorf("no edit list uses animations/%s yet", id)
	}

	return nil
}

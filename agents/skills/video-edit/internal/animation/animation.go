// Package animation draws a scene, the line drawing an animator describes in scene.json, and renders it frame
// by frame in headless Chromium. It drives Chromium over the DevTools pipe: JSON messages, each ending in a NUL
// byte, on file descriptors 3 and 4.
package animation

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed engine.js
var engine string

//go:embed engine.css
var style string

// Cues returns every cue a scene names: each part's cue and until, and each link's cue.
func Cues(scene []byte) (map[string]bool, error) {
	var description struct {
		Parts []struct {
			Cue   string `json:"cue"`
			Until string `json:"until"`
		} `json:"parts"`
		Links []struct {
			Cue string `json:"cue"`
		} `json:"links"`
	}
	err := json.Unmarshal(scene, &description)
	if err != nil {
		return nil, err
	}

	cues := map[string]bool{}
	for _, part := range description.Parts {
		cues[part.Cue] = true
		if part.Until != "" {
			cues[part.Until] = true
		}
	}
	for _, link := range description.Links {
		if link.Cue != "" {
			cues[link.Cue] = true
		}
	}
	delete(cues, "")
	return cues, nil
}

// Page returns the web page that draws the scene, with each cue at its second from the animation's start, and
// clear of the webcam's box, given as fractions of the picture, when webcam is not nil.
func Page(scene []byte, cues map[string]float64, webcam []float64) (string, error) {
	// Encoding the scene again escapes <, > and &, so no text in it can end the script early
	var description any
	err := json.Unmarshal(scene, &description)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(description)
	if err != nil {
		return "", err
	}

	settings, err := json.Marshal(map[string]any{"cues": cues, "webcam": webcam})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("<!doctype html>\n<meta charset=\"utf-8\">\n<style>\n%s</style>\n<body>\n<script>\n%s</script>\n<script>scene(%s, %s)</script>\n",
		style, engine, encoded, settings), nil
}

type Browser struct {
	process *exec.Cmd
	profile string
	writer  *os.File
	reader  *os.File
	buffer  *bufio.Reader
	session string
	lastID  int
	events  []message
	thrown  []string
}

type message struct {
	ID        int             `json:"id"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params"`
	Result    json.RawMessage `json:"result"`
	SessionID string          `json:"sessionId"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Open loads the page in a viewport of width by height CSS pixels, each scale device pixels across, and waits
// for its fonts. A fault the scene finds in itself comes back as the error.
func Open(page string, width, height int, scale float64) (*Browser, error) {
	profile, err := os.MkdirTemp("", "video-edit-chromium-")
	if err != nil {
		return nil, err
	}

	file := filepath.Join(profile, "page.html")
	err = os.WriteFile(file, []byte(page), 0o644)
	if err != nil {
		os.RemoveAll(profile)
		return nil, err
	}

	toChromium, writer, err := os.Pipe()
	if err != nil {
		os.RemoveAll(profile)
		return nil, err
	}

	reader, fromChromium, err := os.Pipe()
	if err != nil {
		os.RemoveAll(profile)
		return nil, err
	}

	// A profile of its own keeps the render away from Raj's running browser
	process := exec.Command("nice", "-n", "19", "chromium", "--headless=new", "--remote-debugging-pipe",
		"--user-data-dir="+profile, "--no-first-run", "--hide-scrollbars", "--mute-audio", "about:blank")
	process.ExtraFiles = []*os.File{toChromium, fromChromium}
	err = process.Start()
	toChromium.Close()
	fromChromium.Close()
	if err != nil {
		os.RemoveAll(profile)
		return nil, fmt.Errorf("chromium failed to start: %w", err)
	}

	browser := &Browser{process: process, profile: profile, writer: writer, reader: reader, buffer: bufio.NewReader(reader)}

	var target struct {
		TargetID string `json:"targetId"`
	}
	err = browser.call("Target.createTarget", map[string]any{"url": "about:blank"}, &target)
	if err != nil {
		browser.Close()
		return nil, err
	}

	var attached struct {
		SessionID string `json:"sessionId"`
	}
	err = browser.call("Target.attachToTarget", map[string]any{"targetId": target.TargetID, "flatten": true}, &attached)
	if err != nil {
		browser.Close()
		return nil, err
	}
	browser.session = attached.SessionID

	steps := []struct {
		method string
		params map[string]any
	}{
		{"Page.enable", nil},
		{"Runtime.enable", nil},
		{"Emulation.setDeviceMetricsOverride", map[string]any{"width": width, "height": height, "deviceScaleFactor": scale, "mobile": false}},
		{"Page.navigate", map[string]any{"url": "file://" + file}},
	}
	for _, step := range steps {
		err = browser.call(step.method, step.params, nil)
		if err != nil {
			browser.Close()
			return nil, err
		}
	}

	for {
		event, err := browser.event()
		if err != nil {
			browser.Close()
			return nil, err
		}

		if event.Method == "Page.loadEventFired" {
			break
		}
	}

	err = browser.evaluate("document.fonts.ready.then(() => {})")
	if err != nil {
		browser.Close()
		return nil, err
	}

	if len(browser.thrown) > 0 {
		browser.Close()
		return nil, fmt.Errorf("%s", strings.Join(browser.thrown, "; "))
	}

	return browser, nil
}

// Frame pauses every animation on the page at the given second and returns the picture as a PNG.
func (b *Browser) Frame(seconds float64) ([]byte, error) {
	err := b.evaluate(fmt.Sprintf(`(() => {
		for (const animation of document.getAnimations()) {
			animation.pause();
			animation.currentTime = %f;
		}
		return new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
	})()`, seconds*1000))
	if err != nil {
		return nil, err
	}

	var screenshot struct {
		Data string `json:"data"`
	}
	err = b.call("Page.captureScreenshot", map[string]any{"format": "png"}, &screenshot)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(screenshot.Data)
}

func (b *Browser) Close() {
	b.writer.Close()
	done := make(chan error, 1)
	go func() { done <- b.process.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		b.process.Process.Kill()
		<-done
	}
	b.reader.Close()
	os.RemoveAll(b.profile)
}

func (b *Browser) evaluate(expression string) error {
	var evaluated struct {
		ExceptionDetails *struct {
			Text      string `json:"text"`
			Exception struct {
				Description string `json:"description"`
			} `json:"exception"`
		} `json:"exceptionDetails"`
	}
	err := b.call("Runtime.evaluate", map[string]any{"expression": expression, "awaitPromise": true}, &evaluated)
	if err != nil {
		return err
	}

	if evaluated.ExceptionDetails != nil {
		return fmt.Errorf("the page failed: %s %s", evaluated.ExceptionDetails.Text, evaluated.ExceptionDetails.Exception.Description)
	}

	return nil
}

func (b *Browser) call(method string, params map[string]any, result any) error {
	b.lastID++
	request := map[string]any{"id": b.lastID, "method": method, "params": params}
	if params == nil {
		request["params"] = map[string]any{}
	}
	if b.session != "" {
		request["sessionId"] = b.session
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = b.writer.Write(append(encoded, 0))
	if err != nil {
		return fmt.Errorf("chromium stopped: %w", err)
	}

	for {
		reply, err := b.read()
		if err != nil {
			return err
		}

		if reply.ID != b.lastID {
			b.events = append(b.events, reply)
			continue
		}

		if reply.Error != nil {
			return fmt.Errorf("chromium %s failed: %s", method, reply.Error.Message)
		}

		if result == nil {
			return nil
		}
		return json.Unmarshal(reply.Result, result)
	}
}

func (b *Browser) event() (message, error) {
	if len(b.events) > 0 {
		event := b.events[0]
		b.events = b.events[1:]
		return event, nil
	}
	return b.read()
}

// read waits at most a minute, so a page that hangs Chromium fails the step instead of stalling it
func (b *Browser) read() (message, error) {
	b.reader.SetReadDeadline(time.Now().Add(time.Minute))
	encoded, err := b.buffer.ReadBytes(0)
	if err != nil {
		return message{}, fmt.Errorf("chromium stopped answering: %w", err)
	}

	var reply message
	err = json.Unmarshal(bytes.TrimSuffix(encoded, []byte{0}), &reply)
	if err != nil {
		return message{}, err
	}

	// A script error would otherwise show only as a wrong picture. The first line of its description holds the
	// message, such as "Error: scene: part "sensor" runs off the stage".
	if reply.Method == "Runtime.exceptionThrown" {
		var exception struct {
			ExceptionDetails struct {
				Text      string `json:"text"`
				Exception struct {
					Description string `json:"description"`
				} `json:"exception"`
			} `json:"exceptionDetails"`
		}
		json.Unmarshal(reply.Params, &exception)
		description, _, _ := strings.Cut(exception.ExceptionDetails.Exception.Description, "\n")
		b.thrown = append(b.thrown, strings.TrimPrefix(description, "Error: "))
	}

	return reply, nil
}

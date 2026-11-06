// Package chromium renders a web page frame by frame in headless Chromium. It drives Chromium over the
// DevTools pipe: JSON messages, each ending in a NUL byte, on file descriptors 3 and 4.
package chromium

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Page struct {
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

// Open loads url in a viewport of width by height CSS pixels, each scale device pixels across, with each
// of properties as a CSS custom property on the root element, and waits for its fonts.
func Open(url string, width, height int, scale float64, properties map[string]string) (*Page, error) {
	profile, err := os.MkdirTemp("", "video-edit-chromium-")
	if err != nil {
		return nil, err
	}

	toChromium, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	reader, fromChromium, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// A profile of its own keeps the render away from Raj's running browser
	process := exec.Command("nice", "-n", "19", "chromium", "--headless=new", "--remote-debugging-pipe",
		"--user-data-dir="+profile, "--no-first-run", "--hide-scrollbars", "--mute-audio",
		"--allow-file-access-from-files", "about:blank")
	process.ExtraFiles = []*os.File{toChromium, fromChromium}
	err = process.Start()
	toChromium.Close()
	fromChromium.Close()
	if err != nil {
		return nil, fmt.Errorf("chromium failed to start: %w", err)
	}

	page := &Page{process: process, profile: profile, writer: writer, reader: reader, buffer: bufio.NewReader(reader)}

	var target struct {
		TargetID string `json:"targetId"`
	}
	err = page.call("Target.createTarget", map[string]any{"url": "about:blank"}, &target)
	if err != nil {
		page.Close()
		return nil, err
	}

	var attached struct {
		SessionID string `json:"sessionId"`
	}
	err = page.call("Target.attachToTarget", map[string]any{"targetId": target.TargetID, "flatten": true}, &attached)
	if err != nil {
		page.Close()
		return nil, err
	}
	page.session = attached.SessionID

	// The properties apply before the page's styles do. An animation that starts running without its cue
	// leaves a faint trace in the frames before the cue, even once paused and seeked.
	var declarations strings.Builder
	for name, value := range properties {
		fmt.Fprintf(&declarations, "%s: %s; ", name, value)
	}
	encodedRule, err := json.Marshal(":root { " + declarations.String() + "}")
	if err != nil {
		page.Close()
		return nil, err
	}

	steps := []struct {
		method string
		params map[string]any
	}{
		{"Page.enable", nil},
		{"Runtime.enable", nil},
		{"Emulation.setDeviceMetricsOverride", map[string]any{"width": width, "height": height, "deviceScaleFactor": scale, "mobile": false}},
		{"Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": fmt.Sprintf("{ const sheet = new CSSStyleSheet(); sheet.replaceSync(%s); document.adoptedStyleSheets = [sheet]; }", encodedRule)}},
		{"Page.navigate", map[string]any{"url": url}},
	}

	for _, step := range steps {
		err = page.call(step.method, step.params, nil)
		if err != nil {
			page.Close()
			return nil, err
		}
	}

	for {
		event, err := page.event()
		if err != nil {
			page.Close()
			return nil, err
		}

		if event.Method == "Page.loadEventFired" {
			break
		}
	}

	err = page.evaluate("document.fonts.ready.then(() => {})")
	if err != nil {
		page.Close()
		return nil, err
	}

	// A script error would otherwise show only as a wrong picture
	if len(page.thrown) > 0 {
		page.Close()
		return nil, fmt.Errorf("%s threw: %s", url, strings.Join(page.thrown, "; "))
	}

	return page, nil
}

// Frame pauses every animation on the page at the given second and returns the picture as a PNG.
func (p *Page) Frame(seconds float64) ([]byte, error) {
	err := p.evaluate(fmt.Sprintf(`(() => {
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
	err = p.call("Page.captureScreenshot", map[string]any{"format": "png"}, &screenshot)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(screenshot.Data)
}

func (p *Page) Close() {
	p.writer.Close()
	done := make(chan error, 1)
	go func() { done <- p.process.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		p.process.Process.Kill()
		<-done
	}
	p.reader.Close()
	os.RemoveAll(p.profile)
}

func (p *Page) evaluate(expression string) error {
	var evaluated struct {
		ExceptionDetails *struct {
			Text      string `json:"text"`
			Exception struct {
				Description string `json:"description"`
			} `json:"exception"`
		} `json:"exceptionDetails"`
	}
	err := p.call("Runtime.evaluate", map[string]any{"expression": expression, "awaitPromise": true}, &evaluated)
	if err != nil {
		return err
	}

	if evaluated.ExceptionDetails != nil {
		return fmt.Errorf("page script failed: %s %s", evaluated.ExceptionDetails.Text, evaluated.ExceptionDetails.Exception.Description)
	}

	return nil
}

func (p *Page) call(method string, params map[string]any, result any) error {
	p.lastID++
	request := map[string]any{"id": p.lastID, "method": method, "params": params}
	if params == nil {
		request["params"] = map[string]any{}
	}
	if p.session != "" {
		request["sessionId"] = p.session
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = p.writer.Write(append(encoded, 0))
	if err != nil {
		return fmt.Errorf("chromium stopped: %w", err)
	}

	for {
		reply, err := p.read()
		if err != nil {
			return err
		}

		if reply.ID != p.lastID {
			p.events = append(p.events, reply)
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

func (p *Page) event() (message, error) {
	if len(p.events) > 0 {
		event := p.events[0]
		p.events = p.events[1:]
		return event, nil
	}
	return p.read()
}

// read waits at most a minute, so a page that hangs Chromium fails the step instead of stalling it
func (p *Page) read() (message, error) {
	p.reader.SetReadDeadline(time.Now().Add(time.Minute))
	encoded, err := p.buffer.ReadBytes(0)
	if err != nil {
		return message{}, fmt.Errorf("chromium stopped answering: %w", err)
	}

	var reply message
	err = json.Unmarshal(bytes.TrimSuffix(encoded, []byte{0}), &reply)
	if err != nil {
		return message{}, err
	}

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
		p.thrown = append(p.thrown, exception.ExceptionDetails.Text+" "+exception.ExceptionDetails.Exception.Description)
	}

	return reply, nil
}

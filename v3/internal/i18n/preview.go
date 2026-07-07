package i18n

import (
	"bufio"
	"os/exec"
	"regexp"
	"sync"

	"github.com/wailsapp/wails/v3/internal/i18n/backend"
)

// previewManager owns the single live-preview dev server (SPEC.md section 12).
// One server at a time, started on demand, torn down on stop/window close.
type previewManager struct {
	be backend.DocsBackend

	mu     sync.Mutex
	cmd    *exec.Cmd
	status string // idle | starting | ready | error
	url    string
	errMsg string
	subs   []chan previewEvent
}

type previewEvent struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

func newPreviewManager(be backend.DocsBackend) *previewManager {
	return &previewManager{be: be, status: "idle"}
}

var devURLRe = regexp.MustCompile(`https?://(?:localhost|127\.0\.0\.1|\[::1\]):\d+`)

// start launches the backend's preview command if not already running.
func (p *previewManager) start(loc backend.Locale) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == "ready" || p.status == "starting" {
		return p.url, nil
	}
	command, ok := p.be.PreviewCommand(loc)
	if !ok {
		p.status = "error"
		p.errMsg = "preview not available for this backend"
		return "", nil
	}
	cmd := exec.Command(command.Name, command.Args...)
	cmd.Dir = command.Dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		p.status = "error"
		p.errMsg = err.Error()
		return "", err
	}
	p.cmd = cmd
	p.status = "starting"
	p.broadcast(previewEvent{Status: "starting"})

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if p.url == "" {
				if m := devURLRe.FindString(line); m != "" {
					p.mu.Lock()
					p.url = m
					p.status = "ready"
					p.broadcast(previewEvent{Status: "ready", URL: m})
					p.mu.Unlock()
				}
			}
		}
		p.mu.Lock()
		if p.status != "error" {
			p.status = "idle"
		}
		p.mu.Unlock()
	}()
	return p.url, nil
}

func (p *previewManager) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}
	p.status = "idle"
	p.url = ""
}

// subscribe returns a channel of preview events for SSE.
func (p *previewManager) subscribe() chan previewEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch := make(chan previewEvent, 8)
	// send current state immediately
	ch <- previewEvent{Status: p.status, URL: p.url, Error: p.errMsg}
	p.subs = append(p.subs, ch)
	return ch
}

func (p *previewManager) broadcast(ev previewEvent) {
	for _, ch := range p.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

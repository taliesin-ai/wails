package commands

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/internal/wake/manifest"
)

// puppertinoRawBaseURL is where the Puppertino release files are fetched from.
// The upstream default branch is "master". The legacy build_assets Taskfile
// template still uses the ".../main/..." path, which returns 404; that template
// is intentionally left unchanged (changing it alters stock Taskfile
// fingerprints the migrator relies on). It is a variable so tests can point
// it at a local server.
var puppertinoRawBaseURL = "https://raw.githubusercontent.com/codedgar/Puppertino/master"

const (
	puppertinoStyleMarker  = `href="/style.css"`
	puppertinoLinkMarker   = `href="/puppertino/puppertino.css"`
	puppertinoLinkTag      = `    <link rel="stylesheet" href="/puppertino/puppertino.css"/>`
	puppertinoButtonBefore = `class="btn"`
	puppertinoButtonAfter  = `class="p-btn p-prim-col"`
	puppertinoMaxBytes     = 16 << 20
)

type ToolPuppertinoOptions struct {
	Frontend string `name:"frontend" description:"Frontend directory containing index.html and public/ (defaults to frontend.directory from wails.yaml, else frontend)"`
	Force    bool   `name:"force" description:"Re-download puppertino.css even when it already exists"`
}

// ToolPuppertino vendors Puppertino CSS into the frontend and wires it into index.html.
// It is idempotent: repeated runs only change files that are not already up to date.
func ToolPuppertino(options *ToolPuppertinoOptions) error {
	DisableFooter = true

	frontend := options.Frontend
	if frontend == "" {
		dir, warning := defaultPuppertinoFrontend()
		if warning != nil {
			fmt.Printf("Warning: %v. Falling back to %q.\n", warning, dir)
		}
		frontend = dir
	}

	publicDir := filepath.Join(frontend, "public", "puppertino")
	cssPath := filepath.Join(publicDir, "puppertino.css")

	if _, err := os.Stat(cssPath); err == nil && !options.Force {
		fmt.Printf("Using bundled Puppertino at %s\n", cssPath)
	} else if err := fetchPuppertino(publicDir, cssPath); err != nil {
		// Matches the Taskfile behaviour: a failed download is not fatal because the template may bundle its own copy.
		fmt.Printf("Warning: could not fetch Puppertino CSS (%v). Proceeding without download.\n", err)
	}

	indexPath := filepath.Join(frontend, "index.html")
	changed, err := patchPuppertinoIndex(indexPath)
	if err != nil {
		return err
	}
	if changed {
		fmt.Printf("Updated %s\n", indexPath)
	}
	return nil
}

// defaultPuppertinoFrontend returns frontend.directory from the manifest when
// one is present, otherwise the conventional "frontend" directory. An invalid
// manifest does not fail the tool: it returns "frontend" together with the
// load error so the caller can warn.
func defaultPuppertinoFrontend() (string, error) {
	root, _, err := manifest.Discover(".")
	if err != nil {
		return "frontend", nil
	}
	loaded, err := manifest.Load(root, "")
	if err != nil {
		return "frontend", fmt.Errorf("could not read %s: %v", manifest.Filename, err)
	}
	if loaded.Config.Frontend.Directory == "" {
		return "frontend", nil
	}
	return filepath.Join(root, loaded.Config.Frontend.Directory), nil
}

func fetchPuppertino(publicDir, cssPath string) error {
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		return err
	}
	css, err := getPuppertinoFile(puppertinoRawBaseURL + "/dist/css/full.css")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cssPath, css, 0o644); err != nil {
		return err
	}
	fmt.Printf("Puppertino CSS downloaded to %s\n", cssPath)

	if license, err := getPuppertinoFile(puppertinoRawBaseURL + "/LICENSE"); err == nil {
		if err := os.WriteFile(filepath.Join(publicDir, "LICENSE"), license, 0o644); err != nil {
			fmt.Printf("Warning: could not write Puppertino LICENSE: %v\n", err)
		}
	} else {
		fmt.Printf("Warning: could not fetch Puppertino LICENSE: %v\n", err)
	}
	return nil
}

func getPuppertinoFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, puppertinoMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > puppertinoMaxBytes {
		return nil, fmt.Errorf("GET %s: response exceeds %d bytes", url, puppertinoMaxBytes)
	}
	return data, nil
}

// patchPuppertinoIndex adds the Puppertino stylesheet link after the first /style.css link
// (when not already present) and switches the default .btn class to Puppertino's primary button.
// It reports whether the file was changed; a missing index.html is not an error.
func patchPuppertinoIndex(indexPath string) (bool, error) {
	info, err := os.Stat(indexPath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return false, err
	}

	html := string(data)
	if !strings.Contains(html, puppertinoLinkMarker) {
		lines := strings.Split(html, "\n")
		out := make([]string, 0, len(lines)+1)
		inserted := false
		for _, line := range lines {
			out = append(out, line)
			if !inserted && strings.Contains(line, puppertinoStyleMarker) {
				out = append(out, puppertinoLinkTag)
				inserted = true
			}
		}
		html = strings.Join(out, "\n")
	}
	html = strings.ReplaceAll(html, puppertinoButtonBefore, puppertinoButtonAfter)

	if html == string(data) {
		return false, nil
	}
	if err := os.WriteFile(indexPath, []byte(html), info.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}

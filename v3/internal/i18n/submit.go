package i18n

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

// submitReq drives the submission flow.
type submitReq struct {
	Locale string `json:"locale"`
	Mode   string `json:"mode"` // export | pr
}

// submitResp is returned for both modes; PR fields are filled only for "pr".
type submitResp struct {
	Mode         string   `json:"mode"`
	ChangedFiles []string `json:"changedFiles"`
	Commands     []string `json:"commands,omitempty"` // copyable manual git/gh commands
	GHAvailable  bool     `json:"ghAvailable"`
	PRURL        string   `json:"prUrl,omitempty"`
	Message      string   `json:"message,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// POST /api/submit
func (a *App) handleSubmit(rw http.ResponseWriter, r *http.Request) {
	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(rw, http.StatusBadRequest, err.Error())
		return
	}
	loc, ok := a.locale(req.Locale)
	if !ok {
		writeErr(rw, http.StatusBadRequest, "unknown locale")
		return
	}

	root := repoRoot(a.be.Root())
	resp := submitResp{Mode: req.Mode, GHAvailable: commandAvailable("gh")}
	resp.ChangedFiles = a.changedTranslationFiles(root, loc.Code)

	branch := "i18n/" + loc.Code + "-update"
	addPaths := submitAddPaths(loc.Code)
	// Manual commands are always provided as the fallback (SPEC.md section 5.5).
	resp.Commands = []string{
		"git checkout -b " + branch,
		"git add " + strings.Join(addPaths, " "),
		fmt.Sprintf("git commit -m %q", "i18n("+loc.Code+"): update translations"),
		"git push -u origin " + branch,
		"gh pr create --fill --web",
	}

	if req.Mode != "pr" {
		resp.Message = "Files are written to your working tree. Use the commands below to submit, or push manually."
		writeJSON(rw, resp)
		return
	}

	// Automated PR via gh (SPEC.md section 5.5).
	if !resp.GHAvailable {
		resp.Error = "gh (GitHub CLI) not found. Use the manual commands below."
		writeJSON(rw, resp)
		return
	}
	if len(resp.ChangedFiles) == 0 {
		resp.Error = "No changed translation files to submit."
		writeJSON(rw, resp)
		return
	}
	build := exec.Command("npm", "run", "build")
	build.Dir = a.be.Root()
	buildOut, err := build.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(buildOut))
		if detail == "" {
			detail = err.Error()
		}
		resp.Error = "npm run build failed: " + detail
		resp.Message = "The automated submit stopped before git changes. Fix the docs build, then use the manual commands below."
		writeJSON(rw, resp)
		return
	}

	// Push to the contributor's fork before opening the PR. `gh pr create`
	// against an unpushed branch fails/prompts, so the push is mandatory here.
	// --fill is non-interactive; gh resolves the fork->upstream head itself.
	gitAdd := append([]string{"git", "-C", root, "add"}, addPaths...)
	steps := [][]string{
		{"git", "-C", root, "checkout", "-b", branch},
		gitAdd,
		{"git", "-C", root, "commit", "-m", "i18n(" + loc.Code + "): update translations"},
		{"git", "-C", root, "push", "-u", "origin", branch},
		{"gh", "pr", "create", "--fill", "--head", branch},
	}
	for _, s := range steps {
		out, err := exec.Command(s[0], s[1:]...).CombinedOutput()
		if err != nil {
			resp.Error = fmt.Sprintf("%s failed: %s", strings.Join(s, " "), strings.TrimSpace(string(out)))
			resp.Message = "The automated submit stopped partway. Finish with the manual commands below."
			writeJSON(rw, resp)
			return
		}
		if s[0] == "gh" {
			resp.PRURL = strings.TrimSpace(string(out))
		}
	}
	resp.Message = "Pull request opened."
	writeJSON(rw, resp)
}

// changedTranslationFiles lists git-changed files under the locale's target dir
// and its cache file.
func (a *App) changedTranslationFiles(root, code string) []string {
	if root == "" {
		return nil
	}
	out, err := exec.Command("git", "-C", root, "status", "--porcelain").Output()
	if err != nil {
		return nil
	}
	var files []string
	prefixes := []string{
		"docs/src/content/docs/" + code + "/",
		"docs/.translation-cache/" + code + ".json",
		"docs/.translation-cache/.snapshots/",
	}
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		for _, p := range prefixes {
			if strings.HasPrefix(path, p) {
				files = append(files, path)
				break
			}
		}
	}
	return files
}

func submitAddPaths(code string) []string {
	return []string{
		"docs/src/content/docs/" + code,
		"docs/.translation-cache/" + code + ".json",
		"docs/.translation-cache/.snapshots",
	}
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

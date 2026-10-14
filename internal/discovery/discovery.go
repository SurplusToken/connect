// Package discovery finds installed tools without reading client configuration.
package discovery

import (
	"encoding/json"
	"howett.net/plist"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type Installation struct {
	Executable    string `json:"executable"`
	Installed     bool   `json:"installed"`
	Version       string `json:"version"`
	VersionStatus string `json:"version_status"`
	Source        string `json:"source"`
	Detail        string `json:"detail"`
	Name          string `json:"name,omitempty"`
	Family        string `json:"family,omitempty"`
}
type Result struct {
	Installation
	Desktop *Installation `json:"desktop,omitempty"`
}
type Options struct {
	Home, OS string
	Dirs     []string
	Run      func(string, []string, string) (string, string)
}

var packages = map[string][]string{
	"codex": {"@openai/codex"}, "claude": {"@anthropic-ai/claude-code"}, "omp": {"@oh-my-pi/pi-coding-agent"}, "opencode": {"@opencode/cli", "opencode-ai"}, "kimi": {"@moonshot-ai/kimi-code"},
}
var semver = regexp.MustCompile(`(?:^|[^0-9A-Za-z])v?([0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?)`)

func version(raw string) string {
	m := semver.FindStringSubmatch(raw)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
func Default() Options {
	home, _ := os.UserHomeDir()
	return Options{Home: home, OS: runtime.GOOS, Dirs: Directories(home, runtime.GOOS, os.Getenv), Run: probe}
}
func Directories(home, goos string, env func(string) string) []string {
	out := []string{}
	add := func(dir string) {
		if filepath.IsAbs(dir) {
			out = append(out, filepath.Clean(dir))
		}
	}
	for _, dir := range filepath.SplitList(env("PATH")) {
		add(dir)
	}
	for _, dir := range []string{filepath.Join(home, ".local", "bin"), filepath.Join(home, "bin"), filepath.Join(home, ".bun", "bin"), filepath.Join(home, ".npm-global", "bin"), filepath.Join(home, ".npm", "bin"), filepath.Join(home, ".volta", "bin"), filepath.Join(home, ".opencode", "bin"), filepath.Join(home, ".kimi", "bin"), filepath.Join(home, ".kimi-code", "bin"), filepath.Join(home, ".local", "share", "mise", "shims"), filepath.Join(home, ".asdf", "shims")} {
		add(dir)
	}
	for _, name := range []string{"VOLTA_HOME", "BUN_INSTALL", "NPM_CONFIG_PREFIX"} {
		if p := env(name); p != "" {
			add(filepath.Join(p, "bin"))
			if goos == "windows" {
				add(p)
			}
		}
	}
	if goos == "windows" {
		if app := env("APPDATA"); app != "" {
			add(filepath.Join(app, "npm"))
		}
		if local := env("LOCALAPPDATA"); local != "" {
			add(filepath.Join(local, "Programs", "nodejs"))
			add(filepath.Join(local, "Microsoft", "WinGet", "Links"))
		}
		for _, name := range []string{"NVM_SYMLINK", "NVM_HOME"} {
			add(env(name))
		}
		if profile := env("USERPROFILE"); profile != "" {
			add(filepath.Join(profile, "scoop", "shims"))
		}
	} else {
		for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/bin"} {
			add(dir)
		}
	}
	nvm := env("NVM_DIR")
	if nvm == "" {
		nvm = filepath.Join(home, ".nvm")
	}
	// Bounded directory discovery only; never source shell startup files.
	for _, pattern := range []string{filepath.Join(nvm, "versions", "node", "*", "bin"), filepath.Join(home, ".local", "share", "fnm", "node-versions", "*", "installation", "bin")} {
		matches, _ := filepath.Glob(pattern)
		for i := len(matches) - 1; i >= 0 && len(matches)-i <= 64; i-- {
			add(matches[i])
		}
	}
	seen := map[string]bool{}
	unique := []string{}
	for _, dir := range out {
		key := dir
		if goos == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			seen[key] = true
			unique = append(unique, dir)
		}
	}
	return unique
}
func find(id string, o Options) (string, string) {
	names := []string{id}
	if o.OS == "windows" {
		names = []string{id + ".exe", id + ".cmd", id + ".bat"}
	}
	for _, dir := range o.Dirs {
		for _, name := range names {
			p := filepath.Join(dir, name)
			st, err := os.Stat(p)
			if err == nil && !st.IsDir() && (o.OS == "windows" || st.Mode()&0111 != 0) {
				return p, dir
			}
		}
	}
	return "", ""
}
func readJSON(path string) (string, string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()
	var p struct{ Name, Version string }
	if json.NewDecoder(io.LimitReader(f, 256<<10)).Decode(&p) != nil {
		return "", ""
	}
	return p.Name, p.Version
}

// Metadata reads only the package associated with the resolved executable.
func metadata(id, path string) string {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		real = path
	}
	candidates := []string{}
	for dir, n := filepath.Dir(real), 0; n < 10; n++ {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		if filepath.Base(parent) == "node_modules" || strings.HasPrefix(filepath.Base(parent), "@") && filepath.Base(filepath.Dir(parent)) == "node_modules" {
			candidates = append(candidates, filepath.Join(dir, "package.json"))
			break
		}
		dir = parent
	}
	for _, name := range packages[id] {
		for _, root := range []string{filepath.Dir(path), filepath.Join(filepath.Dir(path), "..", "lib")} {
			candidates = append(candidates, filepath.Join(root, "node_modules", filepath.FromSlash(name), "package.json"))
		}
	}
	for _, p := range candidates {
		name, v := readJSON(p)
		for _, expected := range packages[id] {
			if name == expected && version(v) == v && v != "" {
				return v
			}
		}
	}
	return ""
}
func Detect(o Options) map[string]Result {
	out := map[string]Result{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, id := range []string{"codex", "claude", "omp", "opencode", "kimi"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			p, source := find(id, o)
			r := Result{Installation: Installation{Executable: p, Installed: p != "", Source: source, VersionStatus: "missing"}}
			if p != "" {
				if v := metadata(id, p); v != "" {
					r.Version = v
					r.VersionStatus = "metadata"
					if id == "kimi" {
						r.Family = "kimi-code"
					}
				}
				if r.Version == "" {
					if id == "codex" {
						r.VersionStatus = "protected"
						r.Detail = "保护中，仅检查安装路径和包版本，不执行 Codex"
					} else if o.OS == "windows" && strings.ToLower(filepath.Ext(p)) != ".exe" {
						r.VersionStatus = "unknown"
						r.Detail = "发现命令包装器，未读取到对应包版本"
					} else {
						run := o.Run
						if run == nil {
							run = probe
						}
						raw, status := run(p, o.Dirs, o.Home)
						r.Version = version(raw)
						r.VersionStatus = status
						if r.Version != "" && status == "ok" {
							r.Detail = "通过 --version 读取"
						} else {
							r.Version = ""
							r.Detail = map[string]string{"timeout": "版本检测超时", "failed": "无法运行版本检测", "unknown": "无法识别版本输出"}[status]
						}
					}
				}
			}
			if id == "codex" && o.OS == "darwin" {
				r.Desktop = desktop(o.Home)
			}
			mu.Lock()
			out[id] = r
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return out
}
func desktop(home string) *Installation {
	for _, name := range []string{"Codex", "ChatGPT"} {
		for _, dir := range []string{filepath.Join(home, "Applications", name+".app"), filepath.Join("/Applications", name+".app")} {
			if st, err := os.Stat(dir); err == nil && st.IsDir() {
				result := &Installation{Name: name, Installed: true, Executable: dir, Source: "应用目录", VersionStatus: "unknown", Detail: "仅读取 Info.plist 应用版本，不启动或修改 Desktop"}
				f, err := os.Open(filepath.Join(dir, "Contents", "Info.plist"))
				if err != nil {
					return result
				}
				var info struct {
					Version string `plist:"CFBundleShortVersionString"`
					Build   string `plist:"CFBundleVersion"`
				}
				raw, readErr := io.ReadAll(io.LimitReader(f, (256<<10)+1))
				if readErr != nil || len(raw) > 256<<10 {
					f.Close()
					return result
				}
				_, err = plist.Unmarshal(raw, &info)
				f.Close()
				if err == nil {
					result.Version = info.Version
					if result.Version == "" {
						result.Version = info.Build
					}
					if result.Version != "" {
						result.VersionStatus = "metadata"
					}
				}
				return result
			}
		}
	}
	return &Installation{VersionStatus: "missing", Detail: "标准应用目录未发现 Codex.app 或 ChatGPT.app"}
}

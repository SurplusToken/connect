package discovery

import (
	"howett.net/plist"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func file(t *testing.T, path, text string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}
func TestMetadataAndCodexNeverExecuted(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	for _, id := range []string{"codex", "claude", "omp", "opencode", "kimi"} {
		file(t, filepath.Join(bin, id), "executable", 0700)
	}
	file(t, filepath.Join(bin, "node_modules", "@openai", "codex", "package.json"), `{"name":"@openai/codex","version":"0.123.0"}`, 0600)
	file(t, filepath.Join(bin, "node_modules", "@anthropic-ai", "claude-code", "package.json"), `{"name":"@anthropic-ai/claude-code","version":"2.5.1"}`, 0600)
	options := Options{Home: root, OS: "darwin", Dirs: []string{bin}, Run: func(path string, dirs []string, home string) (string, string) {
		if filepath.Base(path) == "codex" || filepath.Base(path) == "claude" {
			t.Error("metadata client executed")
		}
		return "tool v1.2.3", "ok"
	}}
	found := Detect(options)
	if found["codex"].Version != "0.123.0" || found["claude"].Version != "2.5.1" || found["omp"].Version != "1.2.3" {
		t.Fatal("incorrect versions")
	}
	os.Remove(filepath.Join(bin, "node_modules", "@openai", "codex", "package.json"))
	found = Detect(options)
	if found["codex"].VersionStatus != "protected" || found["codex"].Version != "" {
		t.Fatal("Codex without metadata must remain protected")
	}
}
func TestPathPriorityPermissionsAndWindowsShim(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "path")
	fallback := filepath.Join(root, "fallback")
	file(t, filepath.Join(first, "claude"), "not executable", 0600)
	file(t, filepath.Join(fallback, "claude"), "executable", 0700)
	p, _ := find("claude", Options{OS: "darwin", Dirs: []string{first, fallback}})
	if p != filepath.Join(fallback, "claude") {
		t.Fatal("non-executable chosen")
	}
	file(t, filepath.Join(first, "claude"), "executable", 0700)
	p, _ = find("claude", Options{OS: "darwin", Dirs: []string{first, fallback}})
	if p != filepath.Join(first, "claude") {
		t.Fatal("PATH priority lost")
	}
	file(t, filepath.Join(first, "opencode.cmd"), "shim", 0600)
	file(t, filepath.Join(first, "node_modules", "opencode-ai", "package.json"), `{"name":"opencode-ai","version":"2.1.0"}`, 0600)
	found := Detect(Options{Home: root, OS: "windows", Dirs: []string{first}, Run: func(string, []string, string) (string, string) {
		t.Fatal("Windows shim interpreted by shell")
		return "", ""
	}})
	if found["opencode"].Version != "2.1.0" {
		t.Fatal("Windows package version missed")
	}
}
func TestWrongPackageNotTrustedAndTimeoutDistinguished(t *testing.T) {
	root := t.TempDir()
	file(t, filepath.Join(root, "kimi"), "executable", 0700)
	file(t, filepath.Join(root, "node_modules", "@moonshot-ai", "kimi-code", "package.json"), `{"name":"unrelated","version":"99.0.0"}`, 0600)
	found := Detect(Options{Home: root, OS: "linux", Dirs: []string{root}, Run: func(string, []string, string) (string, string) { return "", "timeout" }})
	if !found["kimi"].Installed || found["kimi"].Version != "" || found["kimi"].VersionStatus != "timeout" || found["omp"].Installed {
		t.Fatal("timeout or missing status conflated")
	}
}
func TestDirectoriesIncludeGUIFallbacksWithoutRelativePATH(t *testing.T) {
	root := t.TempDir()
	env := func(name string) string {
		if name == "PATH" {
			return "." + string(os.PathListSeparator) + filepath.Join(root, "custom")
		}
		if name == "APPDATA" {
			return filepath.Join(root, "roaming")
		}
		return ""
	}
	dirs := Directories(root, "windows", env)
	joined := strings.Join(dirs, "\n")
	for _, want := range []string{filepath.Join(root, "custom"), filepath.Join(root, ".local", "bin"), filepath.Join(root, "roaming", "npm"), filepath.Join(root, ".bun", "bin")} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s", want)
		}
	}
	for _, dir := range dirs {
		if !filepath.IsAbs(dir) {
			t.Fatal("relative PATH accepted")
		}
	}
}

func TestPassiveDesktopMetadata(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "Applications", "Codex.app")
	file(t, filepath.Join(app, "Contents", "Info.plist"), `<?xml version="1.0"?><plist><dict><key>CFBundleShortVersionString</key><string>9.8.7</string></dict></plist>`, 0600)
	found := Detect(Options{Home: root, OS: "darwin", Run: func(string, []string, string) (string, string) { t.Fatal("Desktop must not execute"); return "", "" }})
	if found["codex"].Desktop == nil || found["codex"].Desktop.Version != "9.8.7" {
		t.Fatal("desktop metadata not read")
	}
}
func TestOutputBound(t *testing.T) {
	var output limitedOutput
	chunk := []byte(strings.Repeat("x", 64000))
	n, err := output.Write(chunk)
	if err != nil || n != len(chunk) || output.b.Len() != 32768 {
		t.Fatal("probe output not bounded")
	}
}

func TestChatGPTDesktopBinaryPlist(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "Applications", "ChatGPT.app")
	data, err := plist.Marshal(map[string]string{"CFBundleShortVersionString": "2026.1009.402", "CFBundleVersion": "9876"}, plist.BinaryFormat)
	if err != nil {
		t.Fatal(err)
	}
	file(t, filepath.Join(app, "Contents", "Info.plist"), string(data), 0600)
	info := desktop(root)
	if info.Name != "ChatGPT" || info.Version != "2026.1009.402" || !info.Installed {
		t.Fatal("ChatGPT desktop metadata missed")
	}
}

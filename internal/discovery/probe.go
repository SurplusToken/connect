package discovery

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type limitedOutput struct {
	mu sync.Mutex
	b  strings.Builder
}

func (b *limitedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	remain := 32*1024 - b.b.Len()
	if remain > 0 {
		if len(p) > remain {
			p = p[:remain]
		}
		b.b.Write(p)
	}
	return n, nil
}
func probe(path string, dirs []string, home string) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	configureProbe(cmd)
	cmd.Dir = home
	cmd.WaitDelay = 200 * time.Millisecond
	env := []string{}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "PATH=") && !strings.HasPrefix(v, "NO_COLOR=") {
			env = append(env, v)
		}
	}
	cmd.Env = append(env, "PATH="+strings.Join(dirs, string(os.PathListSeparator)), "NO_COLOR=1")
	var output limitedOutput
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", "timeout"
	}
	if err != nil {
		return "", "failed"
	}
	raw := output.b.String()
	if version(raw) == "" {
		return "", "unknown"
	}
	return raw, "ok"
}

package clipboard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Copy writes to the system clipboard and reports whether the clipboard tool succeeded.
func Copy(text string) error {
	var candidates [][]string
	switch runtime.GOOS {
	case "darwin":
		candidates = [][]string{{"pbcopy"}}
	case "windows":
		candidates = [][]string{{"clip.exe"}}
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			candidates = append(candidates, []string{"wl-copy"})
		}
		if os.Getenv("DISPLAY") != "" {
			candidates = append(candidates, []string{"xclip", "-selection", "clipboard"}, []string{"xsel", "--clipboard", "--input"})
		}
	default:
		return fmt.Errorf("clipboard copying is unavailable on %s", runtime.GOOS)
	}
	var lastErr error
	for _, args := range candidates {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		err := cmd.Run()
		timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
		cancel()
		if err == nil {
			return nil
		}
		if timedOut {
			lastErr = errors.New("clipboard copy timed out")
			continue
		}
		lastErr = fmt.Errorf("clipboard copy failed: %w", err)
	}
	if lastErr != nil {
		return lastErr
	}
	if runtime.GOOS == "linux" {
		return errors.New("clipboard tool unavailable; install wl-clipboard, xclip, or xsel")
	}
	return errors.New("system clipboard tool unavailable")
}

package engine

import (
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("cmd", "/c", "start", "", url)
	}
	return cmd.Start()
}

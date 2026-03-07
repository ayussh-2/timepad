package ui

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/getlantern/systray"

	"timepad/windows/internal/config"
	"timepad/windows/internal/logger"
)

func RunTray(cfg *config.Config, buf *logger.Buffer, cancel context.CancelFunc) {
	systray.Run(
		func() { onReady(cfg, buf, cancel) },
		func() {},
	)
}

func onReady(cfg *config.Config, buf *logger.Buffer, cancel context.CancelFunc) {
	log.Println("tray: ready")
	systray.SetIcon(iconData)
	systray.SetTitle("Timepad")
	systray.SetTooltip("Timepad")

	mDashboard := systray.AddMenuItem("Open Dashboard", "")
	mLogs := systray.AddMenuItem("View Logs", "")
	systray.AddSeparator()
	mSetServer := systray.AddMenuItem("Set Server URL", cfg.GetServerURL())
	mSetDashboard := systray.AddMenuItem("Set Dashboard URL", cfg.GetDashboardURL())
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	go func() {
		runtime.LockOSThread()
		initDashboard(cfg)
		runtime.UnlockOSThread()
	}()

	go func() {
		for {
			select {
			case <-mDashboard.ClickedCh:
				log.Println("tray: open dashboard clicked")
				ShowDashboard()
			case <-mLogs.ClickedCh:
				log.Println("tray: view logs clicked")
				ShowLogs(buf)
			case <-mSetServer.ClickedCh:
				if url, ok := promptInput("Server URL", "Enter the Timepad API server URL:", cfg.GetServerURL()); ok {
					log.Printf("tray: server URL updated to %s", url)
					cfg.SetServerURL(url)
					mSetServer.SetTooltip(url)
				}
			case <-mSetDashboard.ClickedCh:
				if url, ok := promptInput("Dashboard URL", "Enter the Timepad dashboard URL:", cfg.GetDashboardURL()); ok {
					log.Printf("tray: dashboard URL updated to %s", url)
					cfg.SetDashboardURL(url)
					mSetDashboard.SetTooltip(url)
					NavigateTo(url)
				}
			case <-mQuit.ClickedCh:
				log.Println("tray: quit clicked")
				cancel()
				TerminateDashboard()
				systray.Quit()
				return
			}
		}
	}()
}

// promptInput shows a Windows InputBox via PowerShell and returns the typed
// value. Returns ("", false) if the user cancels or enters an empty string.
func promptInput(title, prompt, defaultVal string) (string, bool) {
	script := fmt.Sprintf(
		`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.Interaction]::InputBox('%s', '%s', '%s')`,
		strings.ReplaceAll(prompt, "'", "'"),
		strings.ReplaceAll(title, "'", "'"),
		strings.ReplaceAll(defaultVal, "'", "'"),
	)
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		log.Printf("tray: promptInput error: %v", err)
		return "", false
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", false
	}
	return result, true
}

package ui

import (
	"context"
	"log"
	"runtime"

	"github.com/getlantern/systray"

	"timepad/windows/internal/config"
)

func RunTray(cfg *config.Config, cancel context.CancelFunc) {
	systray.Run(
		func() { onReady(cfg, cancel) },
		func() {},
	)
}

func onReady(cfg *config.Config, cancel context.CancelFunc) {
	log.Println("tray: ready")
	systray.SetIcon(iconData)
	systray.SetTitle("Timepad")
	systray.SetTooltip("Timepad")

	mDashboard := systray.AddMenuItem("Open Dashboard", "")
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

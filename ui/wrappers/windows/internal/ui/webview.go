package ui

import (
	"fmt"
	"log"
	"sync"
	"syscall"

	webview "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"

	"timepad/windows/internal/config"
)

var wv struct {
	sync.Mutex
	w    webview.WebView
	hwnd windows.HWND
}

var (
	modUser32ui          = windows.NewLazySystemDLL("user32.dll")
	procSetWindowLongPtr = modUser32ui.NewProc("SetWindowLongPtrW")
	procCallWindowProc   = modUser32ui.NewProc("CallWindowProcW")
	procShowWindowWV     = modUser32ui.NewProc("ShowWindow")
	procSetForegroundWV  = modUser32ui.NewProc("SetForegroundWindow")
)

func initDashboard(cfg *config.Config) {
	log.Println("webview: initialising")
	w := webview.New(false)
	if w == nil {
		log.Println("webview: runtime not found")
		return
	}
	defer func() {
		log.Println("webview: destroyed")
		w.Destroy()
	}()

	w.SetTitle("Timepad")
	w.SetSize(1280, 840, webview.HintNone)
	// Phase 1: expose the native bridge object only — no calls yet.
	w.Init(fmt.Sprintf(`
window.TimePadBridge = {
  getDeviceKey: function() { return %q; },
  getPlatform:  function() { return "windows"; }
};
`, cfg.GetDeviceKey()))

	if err := w.Bind("timePadSaveConfig", func(accessToken, refreshToken, deviceKey string) {
		if deviceKey != "" {
			log.Printf("webview: timePadSaveConfig: registration device_key=%q", deviceKey)
		} else {
			log.Printf("webview: timePadSaveConfig: token-restore tokens_present=%v", accessToken != "")
		}
		cfg.SetTokens(accessToken, refreshToken)
		if deviceKey != "" {
			cfg.SetDeviceKey(deviceKey)
			w.Dispatch(func() {
				w.Eval(fmt.Sprintf(`window.TimePadBridge.getDeviceKey = function() { return %q; }`, deviceKey))
			})
		}
		log.Println("webview: config saved")
	}); err != nil {
		log.Printf("webview: bind: %v", err)
	}

	// Phase 2: token-sync IIFE — runs after timePadSaveConfig is bound.
	w.Init(`
(function() {
  try {
    var raw = localStorage.getItem('auth-store');
    if (!raw) return;
    var s = JSON.parse(raw).state;
    if (s && s.accessToken && s.refreshToken) {
      window.timePadSaveConfig(s.accessToken, s.refreshToken, '');
    }
  } catch(_) {}
})();
`)

	log.Printf("webview: navigating to %s", cfg.GetDashboardURL())
	w.Navigate(cfg.GetDashboardURL())

	hwnd := windows.HWND(uintptr(w.Window()))
	log.Printf("webview: hwnd=0x%x", hwnd)
	var oldProc uintptr
	cb := syscall.NewCallback(func(h, msg, wp, lp uintptr) uintptr {
		if msg == 0x0010 { // WM_CLOSE
			log.Println("webview: WM_CLOSE intercepted, hiding window")
			procShowWindowWV.Call(h, 0) // SW_HIDE
			return 0
		}
		ret, _, _ := procCallWindowProc.Call(oldProc, h, msg, wp, lp)
		return ret
	})
	oldProc, _, _ = procSetWindowLongPtr.Call(uintptr(hwnd), ^uintptr(3), cb) // GWLP_WNDPROC = -4

	SetWindowIcon(hwnd)

	wv.Lock()
	wv.w = w
	wv.hwnd = hwnd
	wv.Unlock()

	w.Run()

	wv.Lock()
	wv.w = nil
	wv.hwnd = 0
	wv.Unlock()
}

func ShowDashboard() {
	wv.Lock()
	w := wv.w
	hwnd := wv.hwnd
	wv.Unlock()
	if w == nil {
		log.Println("webview: ShowDashboard called but window not ready")
		return
	}
	log.Println("webview: showing window")
	w.Dispatch(func() {
		procShowWindowWV.Call(uintptr(hwnd), 5) // SW_SHOW
		procSetForegroundWV.Call(uintptr(hwnd))
	})
}

func TerminateDashboard() {
	log.Println("webview: terminating")
	wv.Lock()
	w := wv.w
	wv.Unlock()
	if w != nil {
		w.Terminate()
	}
}

// NavigateTo reloads the webview to the given URL. Safe to call from any goroutine.
func NavigateTo(url string) {
	wv.Lock()
	w := wv.w
	wv.Unlock()
	if w == nil {
		log.Println("webview: NavigateTo called but window not ready")
		return
	}
	w.Dispatch(func() {
		log.Printf("webview: navigating to %s", url)
		w.Navigate(url)
	})
}

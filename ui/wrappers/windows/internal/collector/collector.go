package collector

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"timepad/windows/internal/client"
	"timepad/windows/internal/config"
)

const (
	pollInterval         = 30 * time.Second
	flushInterval        = 60 * time.Second
	maxBufferLen         = 10
	defaultIdleThreshold = 5 * time.Minute
)

// windowsSystemApps is the set of Windows system process names that should
// not be tracked — they produce noise without meaningful activity data.
var windowsSystemApps = map[string]bool{
	"ShellHost":               true,
	"Taskmgr":                 true,
	"conhost":                 true,
	"dwm":                     true,
	"winlogon":                true,
	"csrss":                   true,
	"svchost":                 true,
	"RuntimeBroker":           true,
	"SearchHost":              true,
	"StartMenuExperienceHost": true,
	"ApplicationFrameHost":    true,
	"TextInputHost":           true,
	"ctfmon":                  true,
	"LockApp":                 true,
	"LogonUI":                 true,
	"fontdrvhost":             true,
	"WmiPrvSE":                true,
	"spoolsv":                 true,
	"lsass":                   true,
	"services":                true,
	"Registry":                true,
	"MemCompression":          true,
	"MsMpEng":                 true,
	"SecurityHealthSystray":   true,
	"SecurityHealthService":   true,
	"backgroundTaskHost":      true,
	"dllhost":                 true,
	"SgrmBroker":              true,
	"NisSrv":                  true,
	"smartscreen":             true,
	"UserOOBEBroker":          true,
	"SystemSettings":          true,
	"PhoneExperienceHost":     true,
	"WidgetService":           true,
	"Widgets":                 true,
	"sihost":                  true,
	"taskhostw":               true,
	"wininit":                 true,
	"explorer":                true,
}

// isSystemApp reports whether the given exe name should be filtered out.
func isSystemApp(name string) bool {
	if windowsSystemApps[name] {
		return true
	}
	// All PowerToys utility windows (e.g. PowerToys.PowerLauncher, PowerToys.FancyZones)
	if strings.HasPrefix(name, "PowerToys.") {
		return true
	}
	return false
}

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetForegroundWindow      = modUser32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = modUser32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = modUser32.NewProc("GetWindowThreadProcessId")
	procGetLastInputInfo         = modUser32.NewProc("GetLastInputInfo")
	procGetTickCount             = modKernel32.NewProc("GetTickCount")
)

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

type activeWindow struct {
	exeName string
	title   string
}

type session struct {
	window    activeWindow
	startTime time.Time
}

func getForegroundHWND() windows.HWND {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return windows.HWND(hwnd)
}

func getWindowTitle(hwnd windows.HWND) string {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
}

func getWindowPID(hwnd windows.HWND) uint32 {
	var pid uint32
	procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	return pid
}

func getWindowExe(hwnd windows.HWND) string {
	pid := getWindowPID(hwnd)
	if pid == 0 {
		return ""
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return ""
	}
	name := filepath.Base(windows.UTF16ToString(buf[:size]))
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func getIdleDuration() time.Duration {
	var lii lastInputInfo
	lii.cbSize = uint32(unsafe.Sizeof(lii))
	procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	tick, _, _ := procGetTickCount.Call()
	idleMS := uint32(tick) - lii.dwTime
	return time.Duration(idleMS) * time.Millisecond
}

func snapshotForeground() activeWindow {
	hwnd := getForegroundHWND()
	return activeWindow{exeName: getWindowExe(hwnd), title: getWindowTitle(hwnd)}
}

func Run(ctx context.Context, cfg *config.Config) {
	log.Println("collector: started")

	c := client.New(cfg)
	var buf []client.EventInput
	var cur *session

	pollTick := time.NewTicker(pollInterval)
	flushTick := time.NewTicker(flushInterval)
	defer pollTick.Stop()
	defer flushTick.Stop()

	flush := func() {
		if cur != nil {
			now := time.Now()
			idle := getIdleDuration() > defaultIdleThreshold
			log.Printf("collector: closing session [%s] %q duration=%.0fs idle=%v",
				cur.window.exeName, cur.window.title,
				now.Sub(cur.startTime).Seconds(), idle)
			buf = append(buf, client.EventInput{
				AppName:     cur.window.exeName,
				WindowTitle: cur.window.title,
				StartTime:   cur.startTime,
				EndTime:     now,
				IsIdle:      idle,
			})
			cur.startTime = now
		}
		log.Printf("collector: flush triggered, buffer=%d device_key_set=%v authenticated=%v",
			len(buf), cfg.GetDeviceKey() != "", cfg.GetAccessToken() != "")
		if len(buf) == 0 {
			log.Println("collector: buffer empty, nothing to send")
			return
		}
		if cfg.GetDeviceKey() == "" || cfg.GetAccessToken() == "" {
			log.Printf("collector: not authenticated (device_key=%v access_token=%v), dropping %d events",
				cfg.GetDeviceKey() != "", cfg.GetAccessToken() != "", len(buf))
			buf = buf[:0]
			return
		}
		batch := make([]client.EventInput, len(buf))
		copy(batch, buf)
		buf = buf[:0]
		log.Printf("collector: sending batch of %d events", len(batch))
		go func() {
			if err := c.PostEvents(batch); err != nil {
				log.Printf("collector: flush error: %v", err)
			} else {
				log.Printf("collector: sent %d events OK", len(batch))
			}
		}()
	}

	poll := func() {
		now := time.Now()
		win := snapshotForeground()
		idle := getIdleDuration() > defaultIdleThreshold

		// Skip system processes — they add noise without meaningful data.
		if isSystemApp(win.exeName) {
			return
		}

		if cur == nil {
			log.Printf("collector: first window [%s] %q", win.exeName, win.title)
			cur = &session{window: win, startTime: now}
			return
		}
		if win.exeName == cur.window.exeName && win.title == cur.window.title {
			return
		}
		dur := now.Sub(cur.startTime)
		log.Printf("collector: window change [%s] %q -> [%s] %q (%.0fs, idle=%v)",
			cur.window.exeName, cur.window.title,
			win.exeName, win.title,
			dur.Seconds(), idle)
		buf = append(buf, client.EventInput{
			AppName:     cur.window.exeName,
			WindowTitle: cur.window.title,
			StartTime:   cur.startTime,
			EndTime:     now,
			IsIdle:      idle,
		})
		cur = &session{window: win, startTime: now}
		if len(buf) >= maxBufferLen {
			log.Printf("collector: buffer full (%d), early flush", len(buf))
			flush()
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("collector: shutting down")
			flush()
			return
		case <-pollTick.C:
			poll()
		case <-flushTick.C:
			flush()
		}
	}
}

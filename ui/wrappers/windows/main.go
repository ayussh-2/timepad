package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"timepad/windows/internal/collector"
	"timepad/windows/internal/config"
	"timepad/windows/internal/logger"
	"timepad/windows/internal/ui"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	buf := logger.New()
	log.SetOutput(buf.Writer(os.Stderr))

	log.Println("timepad: starting")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("config: server=%s dashboard=%s", cfg.ServerURL, cfg.DashboardURL)
	if cfg.GetDeviceKey() != "" {
		log.Printf("config: device_key=%s", cfg.GetDeviceKey())
	} else {
		log.Println("config: no device_key — register this device in the dashboard")
	}
	if cfg.GetAccessToken() != "" {
		log.Println("config: access_token present")
	} else {
		log.Println("config: no access_token — log in via the dashboard")
	}

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("timepad: signal received, shutting down")
		cancel()
	}()

	log.Println("timepad: starting collector")
	go collector.Run(ctx, cfg)
	log.Println("timepad: starting tray")
	ui.RunTray(cfg, buf, cancel)
	log.Println("timepad: exited")
}

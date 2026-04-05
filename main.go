package main

import (
	"embed"
	"ffmpeg-wails/backend"
	"fmt"
	"log"
	"os/exec"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := checkFFmpeg(); err != nil {
		log.Printf("WARNING: %v", err)
	}

	app := backend.NewApp()

	if err := wails.Run(&options.App{
		Title:  "FFmpeg-Wails",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.StartupCtx,
		Bind: []interface{}{
			app,
		},
	}); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// checkFFmpeg checks whether ffmpeg and ffprobe are available on the system.
func checkFFmpeg() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return fmt.Errorf("ffprobe not found in PATH")
	}
	return nil
}

package image

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const videoFPS = 5
const videoMaxFrames = 60 // 12 seconds at 5 fps

// HasFFmpeg reports whether ffmpeg is available on the PATH.
func HasFFmpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// RenderVideoFrames downloads and extracts frames from a video file using ffmpeg,
// renders each as an ANSI string, and returns the frames with a fixed inter-frame delay.
// Returns an error if ffmpeg is not installed.
func RenderVideoFrames(raw []byte, cols, rows int) ([]string, []time.Duration, error) {
	if cols <= 0 || rows <= 0 {
		return nil, nil, fmt.Errorf("invalid dimensions: %dx%d", cols, rows)
	}
	if !HasFFmpeg() {
		return nil, nil, fmt.Errorf("video preview requires ffmpeg (not found on PATH)")
	}

	// Write video bytes to a temp input file.
	tmp, err := os.CreateTemp("", "point-tui-video-*")
	if err != nil {
		return nil, nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return nil, nil, fmt.Errorf("writing video temp file: %w", err)
	}
	tmp.Close()

	// Create temp dir for extracted PNG frames.
	frameDir, err := os.MkdirTemp("", "point-tui-frames-*")
	if err != nil {
		return nil, nil, fmt.Errorf("creating frame dir: %w", err)
	}
	defer os.RemoveAll(frameDir)

	cmd := exec.Command("ffmpeg",
		"-i", tmp.Name(),
		"-vf", fmt.Sprintf("fps=%d", videoFPS),
		"-frames:v", fmt.Sprintf("%d", videoMaxFrames),
		"-loglevel", "error",
		filepath.Join(frameDir, "%04d.png"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("ffmpeg failed: %w\n%s", err, out)
	}

	entries, err := os.ReadDir(frameDir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading frame dir: %w", err)
	}
	if len(entries) == 0 {
		return nil, nil, fmt.Errorf("ffmpeg produced no frames")
	}

	delay := time.Second / videoFPS
	frames := make([]string, 0, len(entries))
	delays := make([]time.Duration, 0, len(entries))

	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".png" {
			continue
		}
		pngBytes, err := os.ReadFile(filepath.Join(frameDir, e.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("reading frame %s: %w", e.Name(), err)
		}
		ansi, err := Render(pngBytes, cols, rows)
		if err != nil {
			return nil, nil, fmt.Errorf("rendering frame %s: %w", e.Name(), err)
		}
		frames = append(frames, ansi)
		delays = append(delays, delay)
	}

	return frames, delays, nil
}

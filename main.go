package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "0.0.0.0"
	port = 2222
)

func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "/home/ubuntu/app/ssh-my-id"
	}
	dir := filepath.Dir(exe)
	if dir == "/usr/bin" || dir == "/usr/local/bin" {
		return "/home/ubuntu/app/ssh-my-id"
	}
	return dir
}

func getInfoPanel() string {
	return "\033[1;33m╭─\033[0m \033[1;37mZulfi Fadilah Azhar\033[0m\n" +
		"\033[33m├─\033[0m \033[37mPresident of CodeLabs 2025-2026\033[0m\n" +
		"\033[33m├─\033[0m \033[37mFull Stack Developer\033[0m\n" +
		"\033[33m├─\033[0m \033[37mFocus: RAG, LLM, NLP\033[0m\n" +
		"\033[33m├─\033[0m \033[37mPython | Go | TypeScript\033[0m\n" +
		"\033[33m╰─\033[0m \033[36mgithub.com/ZulfiFazhar\033[0m\n" +
		"\n" +
		"\033[36mShanghai, China\033[0m"
}

func renderNameArt(pngPath string) []string {
	cmd := exec.Command("chafa", "-s", "50x25", "--fg-only", pngPath)
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}
	
	lines := strings.Split(out.String(), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

func extractGifFrames(gifPath string, width int) [][]string {
	// Use chafa to extract all frames from GIF
	cmd := exec.Command("chafa", "-s", fmt.Sprintf("%dx25", width), "--fg-only", "--animate=off", gifPath)
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}
	
	output := out.String()
	
	// Split by frame delimiter - each frame starts after [?25h and ends with [?25l
	// Actually, let's split by the clear screen + hide cursor sequence
	frames := strings.Split(output, "\033[?25l")
	
	var result [][]string
	for _, frame := range frames {
		frame = strings.TrimSpace(frame)
		if frame == "" {
			continue
		}
		// Remove trailing [?25h if present
		frame = strings.TrimSuffix(frame, "\033[?25h")
		frame = strings.TrimSuffix(frame, "\033[?25l")
		
		lines := strings.Split(frame, "\n")
		var frameLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				frameLines = append(frameLines, line)
			}
		}
		if len(frameLines) > 0 {
			result = append(result, frameLines)
		}
	}
	
	return result
}

func streamAnimation(s ssh.Session, gifPath string, namePng string) {
	// Pre-render static content
	nameLines := renderNameArt(namePng)
	if nameLines == nil {
		return
	}
	infoLines := strings.Split(getInfoPanel(), "\n")
	
	// Pre-render all GIF frames
	frames := extractGifFrames(gifPath, 55)
	if len(frames) == 0 {
		return
	}
	
	frameIdx := 0
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			frame := frames[frameIdx%len(frames)]
			output := renderScreen(frame, nameLines, infoLines)
			s.Write([]byte(output))
			
			frameIdx++
			// Infinite loop - frames already contains all GIF frames, just cycle through them
			
		case <-s.Context().Done():
			s.Write([]byte("\033[?25h"))
			return
		}
	}
}

func renderScreen(gifLines []string, nameLines []string, infoLines []string) string {
	var result strings.Builder
	
	// Clear screen, hide cursor, home
	result.WriteString("\033[?25l\033[H\033[2J")
	
	maxLines := len(gifLines)
	if len(nameLines) > maxLines {
		maxLines = len(nameLines)
	}
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}
	
	for i := 0; i < maxLines; i++ {
		// Left: GIF frame
		left := ""
		if i < len(gifLines) {
			left = gifLines[i]
		}
		
		// Right: name art or info
		right := ""
		if i < len(nameLines) {
			right = nameLines[i]
		} else {
			infoIdx := i - len(nameLines)
			if infoIdx >= 0 && infoIdx < len(infoLines) {
				right = infoLines[infoIdx]
			}
		}
		
		if left != "" {
			result.WriteString(left)
		}
		
		if right != "" {
			// Pad left
			for len([]rune(left)) < 58 {
				result.WriteString(" ")
				left += " "
			}
			result.WriteString("  ")
			result.WriteString(right)
		}
		
		if i < maxLines-1 {
			result.WriteString("\r\n")
		}
	}
	
	// Show cursor
	result.WriteString("\033[?25h")
	
	return result.String()
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					exeDir := getExeDir()
					gifPath := filepath.Join(exeDir, "assets", "ascii-animation.gif")
					namePng := filepath.Join(exeDir, "assets", "ascii-art-text-zulfi.png")
					
					streamAnimation(s, gifPath, namePng)
				}
			},
			logging.Middleware(),
		),
		wish.WithHostKeyPath("/home/ubuntu/app/ssh-my-id/id_ed25519"),
	)
	if err != nil {
		fmt.Printf("Gagal membuat server: %v\n", err)
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Server berjalan di %s:%d\n", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			fmt.Printf("Server berhenti: %v\n", err)
		}
	}()

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		fmt.Printf("Gagal mematikan server: %v\n", err)
	}
}

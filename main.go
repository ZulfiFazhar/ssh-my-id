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

func renderGifToAscii(gifPath string, width int) string {
	cmd := exec.Command("chafa", "-s", fmt.Sprintf("%dx25", width), gifPath)
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return out.String()
}

func getBanner() string {
	exeDir := getExeDir()
	pngPath := filepath.Join(exeDir, "assets", "ascii-art-text-zulfi.png")
	gifPath := filepath.Join(exeDir, "assets", "ascii-animation.gif")
	
	// Render PNG name art (left)
	asciiName := renderGifToAscii(pngPath, 80)
	
	// Render GIF as ASCII art (right side, smaller)
	asciiGif := renderGifToAscii(gifPath, 50)
	
	// Combine left-right layout
	nameLines := strings.Split(asciiName, "\n")
	gifLines := strings.Split(asciiGif, "\n")
	
	var result strings.Builder
	result.WriteString("\r\n")
	
	maxLines := len(nameLines)
	if len(gifLines) > maxLines {
		maxLines = len(gifLines)
	}
	
	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(nameLines) {
			left = nameLines[i]
		}
		right := ""
		if i < len(gifLines) {
			right = gifLines[i]
		}
		result.WriteString(left)
		if right != "" {
			// Pad to align
			for len(left) < 90 {
				result.WriteString(" ")
				left += " "
			}
			result.WriteString("  ")
			result.WriteString(right)
		}
		result.WriteString("\r\n")
	}
	
	return result.String()
}

func sshHandler(s ssh.Session) {
	banner := getBanner()
	s.Write([]byte(banner))
	s.Close()
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					banner := getBanner()
					s.Write([]byte(banner))
					next(s)
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

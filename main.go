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

func renderChafa(imagePath string, width int) string {
	cmd := exec.Command("chafa", "-s", fmt.Sprintf("%dx25", width), imagePath)
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

	// Render assets
	gifAscii := renderChafa(gifPath, 55)   // Left: GIF animation
	nameAscii := renderChafa(pngPath, 55)  // Right top: name art
	infoAscii := getInfoPanel()

	gifLines := strings.Split(gifAscii, "\n")
	nameLines := strings.Split(nameAscii, "\n")
	infoLines := strings.Split(infoAscii, "\n")

	// Determine max height
	maxLines := len(gifLines)
	if len(nameLines)+len(infoLines) > maxLines {
		maxLines = len(nameLines) + len(infoLines)
	}

	var result strings.Builder
	result.WriteString("\r\n")

	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(gifLines) {
			left = gifLines[i]
		}

		// Right side: name (top) + info (bottom)
		right := ""
		nameIdx := i
		infoIdx := i - len(nameLines)

		if nameIdx >= 0 && nameIdx < len(nameLines) {
			right = nameLines[nameIdx]
		} else if infoIdx >= 0 && infoIdx < len(infoLines) {
			right = infoLines[infoIdx]
		}

		// Pad left to align with right
		if left != "" {
			result.WriteString(left)
		}

		// Add spacing and right content
		if right != "" {
			// Pad left to ~60 chars before writing right
			for len(left) < 58 {
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

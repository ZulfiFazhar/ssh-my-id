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

// ASCII Art Name
const asciiName = ` ________ __ ______ __ ________ __ __ __ __ ______ __ 
/ | / | / \ / | / | / |/ |/ | / | / \ / | 
$$$$$$$$/ __ __ $$ |/$$$$$$ |$$/ $$$$$$$$/______ ____$$ |$$/ $$ | ______ $$ |____ /$$$$$$ | ________ $$ |____ ______ ______ 
 /$$/ / | / |$$ |$$ |_ $$/ / | $$ |__ / \ / $$ |/ |$$ | / \ $$ \ $$ |__$$ |/ |$$ \ / \ / \ 
 /$$/ $$ | $$ |$$ |$$ | $$ | $$ | $$$$$$ |/$$$$$$$ |$$ |$$ | $$$$$$ |$$$$$$$ | $$ $$ |$$$$$$$$/ $$$$$$$ | $$$$$$ |/$$$$$$ |
 /$$/ $$ | $$ |$$ |$$$$/ $$ | $$$$$/ / $$ |$$ | $$ |$$ |$$ | / $$ |$$ | $$ | $$$$$$$$ | / $$/ $$ | $$ | / $$ |$$ | $$/ 
 /$$/ ____ $$ \__$$ |$$ |$$ | $$ | $$ | /$$$$$$$ |$$ \__$$ |$$ |$$ |/$$$$$$$ |$$ | $$ | $$ | $$ | /$$$$/__ $$ | $$ |/$$$$$$$ |$$ | 
/$$ |$$ $$/ $$ |$$ | $$ | $$ | $$ $$ |$$ $$ |$$ |$$ |$$ $$ |$$ | $$ | $$ | $$ |/$$ |$$ | $$ |$$ $$ |$$ | 
$$$$$$$$/ $$$$$$/ $$/ $$/ $$/ $$/ $$$$$$$/ $$$$$$$/ $$/ $$/ $$$$$$$/ $$/ $$/ $$/ $$/ $$$$$$$$/ $$/ $$/ $$$$$$$/ $$/ `

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
	gifPath := filepath.Join(exeDir, "assets", "ascii-animation.gif")
	
	// Render GIF as ASCII art (left side)
	asciiGif := renderGifToAscii(gifPath, 55)
	
	// Info panel (right side)
	info := `

  \033[1;33m╭─\033[0m \033[1;37mZulfi Fadilah Azhar\033[0m
  \033[33m├─\033[0m \033[37mPresident of CodeLabs 2025-2026\033[0m
  \033[33m├─\033[0m \033[37mFull Stack Developer\033[0m
  \033[33m├─\033[0m \033[37mFocus: RAG, LLM, NLP\033[0m
  \033[33m├─\033[0m \033[37mPython | Go | TypeScript\033[0m
  \033[33m╰─\033[0m \033[36mgithub.com/ZulfiFazhar\033[0m
  
  \033[36mShanghai, China\033[0m`

	// Combine left-right layout
	asciiLines := strings.Split(asciiGif, "\n")
	infoLines := strings.Split(info, "\n")
	
	var result strings.Builder
	
	// Header with ASCII name
	result.WriteString("\r\n")
	result.WriteString("\033[1;36m")
	result.WriteString(asciiName)
	result.WriteString("\033[0m")
	result.WriteString("\r\n")
	
	maxLines := len(asciiLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}
	
	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(asciiLines) {
			left = asciiLines[i]
		}
		right := ""
		if i < len(infoLines) {
			right = infoLines[i]
		}
		if left != "" {
			result.WriteString(left)
		}
		if right != "" {
			// Pad to align with ASCII art
			for len(left) < 60 {
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

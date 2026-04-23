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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
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
	cmd := exec.Command("chafa", "-w", fmt.Sprintf("%d", width), "--fg-only", gifPath)
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
	asciiGif := renderGifToAscii(gifPath, 65)
	
	// Info panel (right side)
	info := fmt.Sprintf(`%s

  %s %s Zulfi Fadilah Azhar
  %s %s President of CodeLabs 2025-2026
  %s %s Full Stack Developer
  %s %s Focus: RAG, LLM, NLP
  %s %s Python | Go | TypeScript
  
  %s github.com/ZulfiFazhar
  %s Shanghai, China`,
		"\x1b[1;36m"+asciiName+"\x1b[0m",
		"\x1b[33m", "\x1b[1;37m",
		"\x1b[33m", "\x1b[1;37m",
		"\x1b[33m", "\x1b[1;37m",
		"\x1b[33m", "\x1b[1;37m",
		"\x1b[33m", "\x1b[1;37m",
		"\x1b[36m",
		"\x1b[36m",
	)
	
	// Combine left-right layout
	asciiLines := strings.Split(asciiGif, "\n")
	infoLines := strings.Split(info, "\n")
	
	var result strings.Builder
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
		result.WriteString(left)
		if right != "" {
			result.WriteString("   ")
			result.WriteString(right)
		}
		result.WriteString("\n")
	}
	
	return result.String()
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return model{content: getBanner()}, []tea.ProgramOption{tea.WithAltScreen()}
}

type model struct {
	content string
}

func (m model) Init() tea.Cmd                           { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m model) View() string                            { return m.content }

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			logging.Middleware(),
		),
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

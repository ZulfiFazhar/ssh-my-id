package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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

// 1. Tentukan Model TUI
type model struct {
	content string
}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string {
	return m.content + "\n\n(Tekan 'q' untuk keluar)\n"
}

// 2. Fungsi untuk menangani koneksi SSH baru
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	str := "--- ZULFI FADILAH AZHAR ---\n" +
		"Full Stack Developer & AI Engineer\n" +
		"President of CodeLabs 2025-2026\n" +
		"Focus: RAG, LLM, NLP"
	
	return model{content: str}, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
	// 3. Setup SSH Server
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

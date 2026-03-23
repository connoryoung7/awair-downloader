package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/connoryoung/awair-downloader/internal/awair"
	"github.com/connoryoung/awair-downloader/internal/tui"
)

func main() {
	token := flag.String("token", os.Getenv("AWAIR_TOKEN"), "Awair API token (or set AWAIR_TOKEN)")
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "error: --token or AWAIR_TOKEN is required")
		flag.Usage()
		os.Exit(1)
	}

	client := awair.NewClient(*token)

	devices, err := client.Devices()
	if err != nil {
		log.Fatalf("fetch devices: %v", err)
	}
	if len(devices) == 0 {
		log.Fatal("no devices found on this account")
	}
	dev := devices[0]

	m := tui.New(client, dev)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalf("tui error: %v", err)
	}
}

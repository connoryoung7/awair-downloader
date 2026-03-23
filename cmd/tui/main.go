package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/connoryoung/awair-downloader/internal/awair"
	"github.com/connoryoung/awair-downloader/internal/tui"
)

const maxCallsPerDay = 400

func main() {
	token := flag.String("token", os.Getenv("AWAIR_TOKEN"), "Awair API token (or set AWAIR_TOKEN)")
	interval := flag.Duration("interval", 5*time.Minute, "Poll interval (minimum 5m, max 400 calls/day)")
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "error: --token or AWAIR_TOKEN is required")
		flag.Usage()
		os.Exit(1)
	}

	minInterval := 5 * time.Minute
	if *interval < minInterval {
		fmt.Fprintf(os.Stderr, "error: --interval must be at least %s (max %d calls/day)\n", minInterval, maxCallsPerDay)
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

	m := tui.New(client, dev, *interval)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalf("tui error: %v", err)
	}
}

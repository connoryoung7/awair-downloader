package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/connoryoung/awair-downloader/internal/awair"
	"github.com/connoryoung/awair-downloader/internal/store"
)

func main() {
	token := flag.String("token", os.Getenv("AWAIR_TOKEN"), "Awair API token (or set AWAIR_TOKEN)")
	fromStr := flag.String("from", "", "Start time for download, e.g. 2026-03-01 (default: 24h ago)")
	toStr := flag.String("to", "", "End time for download, e.g. 2026-03-22 (default: now)")
	interval := flag.Duration("interval", 0, "Poll interval for continuous mode, e.g. 5m (default: one-shot download)")
	dbPath := flag.String("db", "awair.db", "SQLite database file path")
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
	log.Printf("using device: %s (%s/%d)", dev.Name, dev.DeviceType, dev.DeviceID)

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	if *interval > 0 {
		// Continuous poll mode
		log.Printf("polling every %s → %s", *interval, *dbPath)
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-ticker.C:
				r, err := client.Latest(dev.DeviceType, dev.DeviceID)
				if err != nil {
					log.Printf("fetch error: %v", err)
					continue
				}
				if err := s.Insert(r); err != nil {
					log.Printf("insert error: %v", err)
					continue
				}
				temp, _ := r.Sensor("temp")
				co2, _ := r.Sensor("co2")
				log.Printf("recorded: score=%.0f temp=%.1f co2=%.0f", r.Score, temp, co2)
			case <-sig:
				log.Println("shutting down")
				return
			}
		}
	}

	// One-shot download mode
	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now

	if *fromStr != "" {
		from, err = time.ParseInLocation("2006-01-02", *fromStr, time.Local)
		if err != nil {
			log.Fatalf("invalid --from: %v", err)
		}
	}
	if *toStr != "" {
		to, err = time.ParseInLocation("2006-01-02", *toStr, time.Local)
		if err != nil {
			log.Fatalf("invalid --to: %v", err)
		}
		to = to.Add(24*time.Hour - time.Second) // inclusive end of day
	}

	log.Printf("downloading %s → %s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	readings, err := client.RawData(dev.DeviceType, dev.DeviceID, from, to)
	if err != nil {
		log.Fatalf("fetch data: %v", err)
	}

	for _, r := range readings {
		if err := s.Insert(&r); err != nil {
			log.Printf("insert error: %v", err)
		}
	}

	log.Printf("done: %d readings saved to %s", len(readings), *dbPath)
}

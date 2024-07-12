package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-co-op/gocron/v2"
)

type IPAddress struct {
	IP string `json:"ip"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("error loading config: %w", err))
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	slog.Info("verifying Cloudflare API token")
	token, err := api.VerifyAPIToken(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to verify API token: %s", err.Error()))
		return
	}

	slog.Info(fmt.Sprintf("API token verified: %s", token.Status))

	// first run
	handler(ctx, api, cfg.DryRun, cfg.ZoneName, cfg.RecordName)()

	if cfg.Frequency <= 0 {
		slog.Debug("no frequency set, exiting")
		return
	}

	interval := time.Duration(cfg.Frequency) * time.Minute
	slog.Debug(fmt.Sprintf("scheduling job every %s minutes", interval))

	_, err = s.NewJob(gocron.DurationJob(interval), gocron.NewTask(handler(ctx, api, cfg.DryRun, cfg.ZoneName, cfg.RecordName)))
	if err != nil {
		log.Fatal(err)
	}

	s.Start()

	// Wait for CTRL-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	// We block here until a CTRL-C / SigInt is received
	// Once received, we exit and the server is cleaned up
	<-sigChan

	slog.Info("shutting down")
	err = s.Shutdown()
	if err != nil {
		log.Fatal(err)
	}
}

func handler(ctx context.Context, api *cloudflare.API, dryRun bool, zoneName, recordName string) func() {
	return func() {
		resp, err := http.Get("https://api.ipify.org?format=json")
		if err != nil {
			slog.Error(fmt.Sprintf("failed to send ipify request: %s", err.Error()))
			return
		}
		defer resp.Body.Close()

		var ip IPAddress
		if err := json.NewDecoder(resp.Body).Decode(&ip); err != nil {
			slog.Error(fmt.Sprintf("failed to decode ipify response: %s", err.Error()))
			return
		}

		slog.Info(fmt.Sprintf("ipify returned IP address: %s", ip.IP))

		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			log.Fatal(err)
		}

		slog.Info("getting DNS records")
		records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Type: "A"})
		if err != nil {
			return
		}

		slog.Debug(fmt.Sprintf("found %d records", len(records)))
		var target *cloudflare.DNSRecord
		for _, record := range records {
			if record.Name == recordName {
				target = &record
				break
			}
		}

		if target == nil {
			slog.Error("configured target record not found")
			return
		}

		slog.Debug(fmt.Sprintf("target record: %s, %s", target.Name, target.Content))

		if target.Content == ip.IP {
			slog.Info("target record already up to date, skipping update")
			return
		}

		if dryRun {
			slog.Info("dry run enabled, skipping update")
			return
		}

		_, err = api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{ID: target.ID, Content: ip.IP})
		if err != nil {
			slog.Error(fmt.Sprintf("failed to update target DNS record content: %s", err.Error()))
			return
		}

		slog.Info("IP address set on target DNS record")
	}
}

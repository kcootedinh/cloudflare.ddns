package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/go-co-op/gocron/v2"
)

type IPAddress struct {
	IP string `json:"ip"`
}

func main() {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := loadConfig()
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

	if cfg.Frequency <= 0 {
		handler(ctx, api, cfg.DryRun, cfg.ZoneName, cfg.RecordName)()
		return
	}

	_, err = s.NewJob(gocron.DurationJob(time.Duration(cfg.Frequency)*time.Minute), gocron.NewTask(handler(ctx, api, cfg.DryRun, cfg.ZoneName, cfg.RecordName)))
	if err != nil {
		log.Fatal(err)
	}

	s.Start()

	select {
	case <-time.After(time.Minute):
	}

	err = s.Shutdown()
	if err != nil {
		log.Fatal(err)
	}
}

func handler(ctx context.Context, api *cloudflare.API, dryRun bool, zoneName, recordName string) func() {
	return func() {
		resp, err := http.Get("https://api.ipify.org?format=json")
		if err != nil {
			slog.Error(fmt.Sprintf("failed to get IP address: %s", err.Error()))
			return
		}
		defer resp.Body.Close()

		var ip IPAddress
		if err := json.NewDecoder(resp.Body).Decode(&ip); err != nil {
			slog.Error(fmt.Sprintf("failed to decode IP address: %s", err.Error()))
			return
		}

		slog.Info(fmt.Sprintf("IP address: %s", ip.IP), slog.Time("time", time.Now()))

		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			log.Fatal(err)
		}

		records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Type: "A"})
		if err != nil {
			return
		}

		var target *cloudflare.DNSRecord
		for _, record := range records {
			if record.Name == recordName {
				target = &record
				break
			}
		}

		if target == nil {
			slog.Error("target record not found")
			return
		}

		slog.Info(fmt.Sprintf("target record: %s, %s", target.Name, target.Content))

		if target.Content == ip.IP {
			slog.Info("IP address has not changed")
			return
		}

		if dryRun {
			slog.Info("dry run enabled, skipping update")
			return
		}

		_, err = api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{ID: target.ID, Content: ip.IP})
		if err != nil {
			slog.Error(fmt.Sprintf("failed to update IP address: %s", err.Error()))
			return
		}

		slog.Info("IP address updated")
	}
}

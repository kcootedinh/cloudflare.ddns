package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	DryRun        bool
	Frequency     int
	CloudflareKey string
	ZoneName      string
	RecordName    string
}

func loadConfig() (Config, error) {
	dryRun, ok := os.LookupEnv("DRY_RUN")
	if !ok {
		dryRun = "false"
	}
	dryRunBool, err := strconv.ParseBool(dryRun)

	cloudflareKey, ok := os.LookupEnv("CLOUDFLARE_API_TOKEN")
	if !ok {
		return Config{}, errors.New("CLOUDFLARE_API_TOKEN is required")
	}

	frequency, ok := os.LookupEnv("POLL_FREQUENCY")
	if !ok {
		frequency = "0"
	}

	frequencyInt, err := strconv.ParseInt(frequency, 10, 32)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to parse POLL_FREQUENCY, %s", err.Error()))
		frequencyInt = 0
	}

	zoneName, ok := os.LookupEnv("ZONE_NAME")
	if !ok {
		return Config{}, errors.New("ZONE_NAME is required")
	}

	recordName, ok := os.LookupEnv("RECORD_NAME")
	if !ok {
		return Config{}, errors.New("RECORD_NAME is required")
	}

	return Config{
		DryRun:        dryRunBool,
		Frequency:     int(frequencyInt),
		CloudflareKey: cloudflareKey,
		ZoneName:      zoneName,
		RecordName:    recordName,
	}, nil
}

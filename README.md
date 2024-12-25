# Overview

A Go app written to poll the WAN IP address of a local network and then update a Cloudflare managed DNS record. Basically a Cloudflare DDNS updater.

I wrote this to use on a NAS as Cloudflare wasn't a supported DDNS provider of either Synology or Unifi.

## Quick start
`docker run -e CLOUDFLARE_API_TOKEN=<token> -e ZONE_NAME=<dns zone> -e RECORD_NAME=<A record to update> ghcr.io/kcootedinh/cloudflare.ddns:latest`

## Optional
You can specify the log level with the environment var `LOG_LEVEL`. These are the Golang slog log levels. Default is level 0 and equates to `INFO`. `DEBUG` is "-4".

The quick start above will run the update a single time. If you want to run it at a regular interval, you can specify the interval in minutes with the environment var `JOB_FREQUENCY`. For example, to run every 10 minutes, add `-e JOB_FREQUENCY=10`.

You can also specify `DRY_RUN` to just log the changes that would be made without actually making them. This is useful for testing. Specify any value, e.g. `DRY_RUN=true`. Note this will still validate the Cloudflare token and get the current value of the DNS record to check if a change is required.

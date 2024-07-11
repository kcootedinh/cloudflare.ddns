# syntax=docker/dockerfile:1

FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudflare-ddns-updater

FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM gcr.io/distroless/static-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /cloudflare-ddns-updater /cloudflare-ddns-updater

USER nonroot:nonroot

ENTRYPOINT ["/cloudflare-ddns-updater"]
## Compile executable
FROM alpine:3 AS compiler

RUN apk add --no-cache go

# Set the working directory inside the container
WORKDIR /app/NginxUptimeGo

# Copy the entire project directory into the container
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" .

## Compress executable
FROM rexezugedockerutils/upx AS upx

FROM debian:stable-slim AS compressor

COPY --from=upx /upx /usr/local/bin/upx

COPY --from=0 /app/NginxUptimeGo/nginx-uptime-go /NginxUptime-Go

RUN upx --best --lzma /NginxUptime-Go

# Final stage
FROM busybox:stable-musl

COPY --from=compressor /NginxUptime-Go /NginxUptime-Go

# Set default command
ENTRYPOINT ["/NginxUptime-Go"]

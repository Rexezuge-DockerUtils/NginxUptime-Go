## Compile executable
FROM alpine:3 AS compiler

RUN apk add --no-cache go

# Set the working directory inside the container
WORKDIR /app/NginxUptimeGo

# Copy the entire project directory into the container
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" .

## Compress executable
FROM debian:12 AS compressor

WORKDIR /tmp

COPY --from=0 /app/NginxUptimeGo/nginx-uptime-go /NginxUptime-Go

# Install dependencies
RUN apt-get update \
 && apt-get install -y --no-install-recommends build-essential curl ca-certificates

ARG UPX_VERSION=5.0.2

RUN curl -L https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz -o /tmp/upx.tar.xz \
 && tar -xf /tmp/upx.tar.xz -C /tmp \
 && mv /tmp/upx-${UPX_VERSION}-amd64_linux/upx /usr/local/bin/upx

RUN upx --best --lzma /NginxUptime-Go

# Final stage
FROM busybox:stable-musl

COPY --from=compressor /NginxUptime-Go /NginxUptime-Go

# Set default command
ENTRYPOINT ["/NginxUptime-Go"]

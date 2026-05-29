# Build a static partition-vacuum binary and ship it on a minimal base.
# CGO-free (disk stats via syscalls), so the binary runs anywhere.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.version=${VERSION}" -o /out/partition-vacuum .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/partition-vacuum /usr/bin/partition-vacuum
# No args: reads /etc/partition-vacuum/*.toml (same default path the AUR
# systemd unit uses). Mount a config.toml there.
ENTRYPOINT ["/usr/bin/partition-vacuum"]

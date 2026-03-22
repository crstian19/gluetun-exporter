FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
      -X main.Version=${VERSION} \
      -X main.GitCommit=${GIT_COMMIT} \
      -X main.BuildDate=${BUILD_DATE}" \
    -o gluetun-exporter .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/gluetun-exporter .

EXPOSE 9586

USER nobody

ENTRYPOINT ["/app/gluetun-exporter"]

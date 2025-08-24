# Stage 1
FROM golang:1.25.0 AS builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o server ./cmd

# Stage 2
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot:nonroot
EXPOSE 3000
ENTRYPOINT ["./server", "start"]

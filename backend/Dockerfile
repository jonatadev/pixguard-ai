FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pixguard-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /pixguard-api /pixguard-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/pixguard-api"]

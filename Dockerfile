FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/app/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/app /app/app

COPY --from=builder /app/configs ./configs
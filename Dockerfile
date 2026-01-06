FROM golang:1.23-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/music-bot-v2 ./cmd/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/music-bot-v2 /app/music-bot-v2

EXPOSE 8080

ENTRYPOINT ["/app/music-bot-v2"]

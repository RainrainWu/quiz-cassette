# Build Base Container
FROM golang:1.16-stretch AS base

ENV CGO_ENABLED=0
WORKDIR /cassette

COPY . .

RUN go mod download && \
    go build -o ./build/cassette ./cmd && \
    chmod +x ./build/cassette

# Application Container
FROM alpine:3.13

COPY --from=base /cassette/build/cassette /app/cassette

RUN adduser -S rain && chown -R rain /app
USER rain
CMD ["./app/cassette"]
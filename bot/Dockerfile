FROM golang:1.23.4 AS builder

WORKDIR /bot

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 go build -o /download_bot  ./cmd

FROM alpine:3.21

COPY --from=builder /download_bot /download_bot

CMD ["/download_bot"]
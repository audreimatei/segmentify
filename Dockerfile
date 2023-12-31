FROM golang:1.21.0-alpine3.18

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/segmentify

CMD ["./app"]

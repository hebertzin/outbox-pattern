FROM golang:1.22.0-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o transaction-service ./cmd/main.go

EXPOSE 8080

CMD ["./transaction-service"]
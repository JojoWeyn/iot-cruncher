FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main /main


EXPOSE 8080

CMD ["/main"]

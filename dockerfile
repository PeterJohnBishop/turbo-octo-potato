FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server-ssh

FROM alpine:latest

WORKDIR /app

COPY --from=builder /server-ssh .

EXPOSE 2222

CMD ["./server-ssh"]
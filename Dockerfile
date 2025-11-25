FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .


COPY --from=builder /app/migration ./migration

EXPOSE 8080

CMD ["./server"]

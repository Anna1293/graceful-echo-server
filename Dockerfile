FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/graceful-echo-server .

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /out/graceful-echo-server /app/graceful-echo-server

EXPOSE 8080 8443

CMD ["/app/graceful-echo-server"]

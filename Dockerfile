FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o icoconv .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/icoconv .
EXPOSE 8080
CMD ["./icoconv"]

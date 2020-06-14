FROM golang:latest as builder
WORKDIR /app
ENV GOPROXY https://goproxy.io
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minitube .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache bash
# COPY --from=builder /app/out ./out
COPY --from=builder /app/minitube .
COPY --from=builder /app/healthcheck /usr/local/bin/
COPY --from=builder /app/wait-for-it /usr/local/bin/
HEALTHCHECK --start-period=32s --interval=32s --timeout=2s --retries=3 CMD healthcheck
EXPOSE 80
CMD ["./minitube"]
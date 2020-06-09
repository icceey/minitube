FROM golang:latest as builder
WORKDIR /app
ENV GOPROXY https://goproxy.io
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minitube .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/minitube .
EXPOSE 80
ENTRYPOINT ["./minitube"]
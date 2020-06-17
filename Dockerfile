FROM golang:latest as builder
WORKDIR /app
ENV GOPROXY https://goproxy.io
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minitube .

FROM node:14.4.0-alpine as front
WORKDIR /app
RUN npm config set registry https://registry.npm.taobao.org
RUN wget https://github.com/YuJuncen/minitube-fnt/archive/master.zip
RUN unzip master.zip
WORKDIR /app/minitube-fnt-master
RUN yarn install
RUN npx next build
RUN npx next export

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache bash
COPY --from=front /app/minitube-fnt-master/out ./out
COPY --from=builder /app/minitube .
COPY --from=builder /app/healthcheck /usr/local/bin/
COPY --from=builder /app/wait-for-it /usr/local/bin/
HEALTHCHECK --start-period=32s --interval=32s --timeout=2s --retries=3 CMD healthcheck
EXPOSE 80
CMD ["./minitube"]
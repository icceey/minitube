#!/bin/sh

echo 'test start...'

docker run --rm -d --name minitube-mysql-test -p 3306:3306 \
    -e MYSQL_USER=minitube \
    -e MYSQL_PASSWORD=minitube \
    -e MYSQL_DATABASE=minitube \
    -e MYSQL_ROOT_PASSWORD=minitube \
    mysql

docker run --rm -d --name minitube-redis-test -p 6379:6379 \
    redis:alpine redis-server --requirepass minitube

docker run --rm -d --name minitube-live-test -p 8090:8090 \
    -e REDIS_ADDR=localhost:6379 \
    -e REDIS_PASSWORD=minitube \
    icceey/livego 

export MYSQL_USER=minitube
export MYSQL_PASSWORD=minitube
export MYSQL_DATABASE=minitube
export MYSQL_ROOT_PASSWORD=minitube
export REDIS_PASSWORD=minitube
export JWT_SECRET_KEY=minitube
export MYSQL_ADDR=localhost:3306
export REDIS_ADDR=localhost:6379
export LIVE_ADDR=localhost:8090

# wait for mysql container initialize.
sleep 15s

echo 'run go test'
go test -race -v -count=1 ./store
go test -race -v -count=1 ./api

unset MYSQL_USER
unset MYSQL_PASSWORD
unset MYSQL_DATABASE
unset MYSQL_ROOT_PASSWORD
unset REDIS_PASSWORD
unset JWT_SECRET_KEY
unset MYSQL_ADDR
unset REDIS_ADDR
unset LIVE_ADDR

echo 'Stopping docker container...'
docker stop minitube-live-test
docker stop minitube-redis-test
docker stop minitube-mysql-test

#!/bin/sh
set -eo pipefail

host="$(hostname -i || echo '127.0.0.1')"

if ping="$(redis-cli -h "$host" -a "$REDIS_PASSWORD" ping 2>&1 | grep PONG)" && [ "$ping" = 'PONG' ]; then
	exit 0
fi

exit 1
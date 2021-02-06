#!/bin/sh
export PORT=80
export GIN_MODE=debug
export REDIS_ADDR=redis:6379
export REDIS_PASSWD=somepasswd
export REDIS_CACHE_EXPIRATION=3600
export MAX_VISIT_COUNT=1000

docker-compose up
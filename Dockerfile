FROM golang:1.12.5-alpine3.9 AS builder

WORKDIR /app
COPY . /app

RUN apk add --no-cache git gcc
RUN apk add libc-dev

RUN go build rediseen

# Unit Test
RUN REDISEEN_REDIS_URI=redis://:@localhost:6400 REDISEEN_KEY_PATTERN_EXPOSED="^key:[.]*" \
    REDISEEN_TEST_MODE=true REDISEEN_DB_EXPOSED=0-5 \
    go test -cover rediseen


# For smaller image size
# see https://medium.com/@gdiener/how-to-build-a-smaller-docker-image-76779e18d48a
FROM alpine:3.9
WORKDIR /app
COPY --from=builder /app/rediseen ./rediseen

# To allow the service to be accessible outside the container
ENV REDISEEN_HOST=0.0.0.0

EXPOSE 8000

ENV PATH=$PATH:${pwd}

CMD ["./rediseen", "start"]

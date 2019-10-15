FROM golang:1.12.5-alpine3.9 AS builder

WORKDIR /app
COPY . /app

RUN apk add --no-cache git gcc
RUN apk add libc-dev

RUN go build rediseen

# Unit Test
ENV REDISEEN_REDIS_URI=redis://:@localhost:6400
ENV REDISEEN_KEY_PATTERN_EXPOSED="^key:[.]*"
ENV REDISEEN_TEST_MODE=true
ENV REDISEEN_DB_EXPOSED=0-5
RUN go test -cover .

ENV REDISEEN_REDIS_URI=
ENV REDISEEN_KEY_PATTERN_EXPOSED=
ENV REDISEEN_TEST_MODE=
ENV REDISEEN_DB_EXPOSED=


# For smaller image size
# see https://medium.com/@gdiener/how-to-build-a-smaller-docker-image-76779e18d48a
FROM alpine:3.9
WORKDIR /app
COPY --from=builder /app/rediseen ./rediseen

CMD ["./rediseen"]
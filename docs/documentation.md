# Rediseen Documentation

- [Installation](#installation)
- [Configuration](#configuration)
- [How to Start the Service](#how-to-start-the-service)
- [How to Consume the Service](#how-to-consume-the-service)
- [API Authentication](#api-authentication)

## Installation 

You can choose to install `Rediseen` by using `Homebrew`, building from source, or using Docker.

### Install Using `Homebrew`

You can use [Homebrew](https://brew.sh/) to install `Rediseen`, no matte you are using `macOS`, or `Linux`/
`Windows 10 Subsystem for Linux` ([how to install Homebrew](https://docs.brew.sh/Installation)).

```bash
brew install XD-DENG/rediseen/rediseen

rediseen help
```

### Build from source (with Go 1.12 or above installed)

You can also build `Rediseen` from source.

```bash
git clone https://github.com/XD-DENG/rediseen.git
cd rediseen
go build . # executable binary file "rediseen" will be created

./rediseen help
```

### Run with Docker
```bash
docker run \
    -e REDISEEN_REDIS_URI="redis://:@[YOUR REDIS HOST]:6379" \
    -e REDISEEN_DB_EXPOSED=0 \
    -e REDISEEN_KEY_PATTERN_EXPOSED="^key:([0-9a-z]+)" \
    -p 8000:8000 \
    xddeng/rediseen:latest
```


## Configuration

Configuration is done via **environment variables**.

| Item | Description | Remark |
| --- | --- | --- |
| `REDISEEN_REDIS_URI` | URI of your Redis database, e.g. `redis://:@localhost:6379` | Compulsory |
| `REDISEEN_HOST` | Host of the service. Host will be `localhost` if `REDISEEN_HOST` is not explicitly set. Set to `0.0.0.0` if you want to expose your service to the world. | Optional |
| `REDISEEN_PORT` | Port of the service. Default port is 8000. | Optional |
| `REDISEEN_DB_EXPOSED` | Redis logical database(s) to expose.<br><br>E.g., `0`, `0;3;9`, `0-9;15`, or `*` (expose all logical databases) | Compulsory |
| `REDISEEN_KEY_PATTERN_EXPOSED` | Regular expression pattern, representing the name pattern of keys that you intend to expose.<br><br>For example, `user:([0-9a-z/.]+)\|^info:([0-9a-z/.]+)` exposes keys like `user:1`, `user:x1`, `testuser:1`, `info:1`, etc. |  |
| `REDISEEN_KEY_PATTERN_EXPOSE_ALL` | If you intend to expose ***all*** your keys, set `REDISEEN_KEY_PATTERN_EXPOSE_ALL` to `true`. | `REDISEEN_KEY_PATTERN_EXPOSED` can only be empty (or not set) if you have set `REDISEEN_KEY_PATTERN_EXPOSE_ALL` to `true`. |
| `REDISEEN_API_KEY` | API Key for authentication. Authentication is only enabled when `REDISEEN_API_KEY` is set and is not "".<br><br>Once it is set, client must add the API key into HTTP header as field `X-API-KEY` in order to access the API.<br><br>Note this authentication is only considered secure if used together with other security mechanisms such as HTTPS/SSL [1]. | Optional |
| `REDISEEN_TEST_MODE` | Set to `true` to skip Redis connection validation for unit tests. | For Dev Only |


## How to Start the Service

Run command below,

```bash
rediseen start
```

Then you can access the service at
- `http://<your server address>:<REDISEEN_PORT>/<redis DB>`
- `http://<your server address>:<REDISEEN_PORT>/<redis DB>/<key>`
- `http://<your server address>:<REDISEEN_PORT>/<redis DB>/<key>/<index or value or member>`

If you would like to run the service in daemon mode, apply flag `-d`.

```bash
# run service in daemon mode
rediseen start -d

# stop service running in daemon mode
rediseen stop
```


## How to Consume the Service

### 1 `/<redis DB>`

This endpoint will return response in which you can get
- the number of keys which are exposed
- keys exposed and their types (**only up to 1000 keys will be showed**)

A sample response follows below

```
{
    "count": 3,
    "total": 3,
    "keys": [
        {
            "key": "key:1",
            "type": "string"
        },
        {
            "key": "key:5",
            "type": "hash"
        },
        {
            "key": "key:100",
            "type": "zset"
        }
    ]
}
```

### 2 `/<redis DB>/<key>`

| Data Type | Underlying Redis Command |
| --- | --- |
| STRING | `GET(key)` |
| LIST   | `LRANGE(key, 0, -1)` |
| SET    | `SMEMBERS(key)` |
| HASH   | `HGETALL(key)` |
| ZSET   | `ZRANGE(key, 0, -1)` |


### 3 `/<redis DB>/<key>/<index or value or member>`

| Data Type | Usage | Return Value |
| --- | --- | --- |
| STRING | `/<redis DB>/<key>/<index>`  | `<index>`-th character in the string |
| LIST   | `/<redis DB>/<key>/<index>` | `<index>`-th element in the list |
| SET    | `/<redis DB>/<key>/<member>` | if `<member>` is member of the set |
| HASH   | `/<redis DB>/<key>/<field>` | value of hash `<field>` in the hash |
| ZSET   | `/<redis DB>/<key>/<memeber>` | index of `<member>` in the sorted set |

### 4 `/info`

It returns ALL results from [Redis `INFO` command](https://redis.io/commands/info) as a nicely-formatted JSON object.

### 5 `/info/<info_section>`

It returns results from [Redis `INFO <SECTION>` command](https://redis.io/commands/info) as a nicely-formatted JSON object.

Supported `info_section` values can be checked by querying `/info`. They vary according to your Redis version.


## API Authentication

API Key authentication is supported.

To enable authentication, simply set environment variable `REDISEEN_API_KEY` and the value would be the key.
Once it's set, client will have to add the API key as `X-API-KEY` in their HTTP header in order to access anything
meaningful, otherwise 401 error (`Unauthorized`) will be returned.

For example,

```bash
export REDISEEN_REDIS_URI="redis://:@localhost:6379"
export REDISEEN_DB_EXPOSED=0
export REDISEEN_KEY_PATTERN_EXPOSED="^key:([0-9a-z]+)"
export REDISEEN_API_KEY="demo_key" # Set REDISEEN_API_KEY to enforce API Key Authentication

# Start the service and run in background
rediseen start -d

# REJECTED: No X-API-KEY is given in HTTP header
curl -s http://localhost:8000/0 | jq
{
  "error": "unauthorized"
}

# REJECTED: Wrong X-API-KEY is given in HTTP header
curl -s -H "X-API-KEY: wrong_key" http://localhost:8000/0 | jq
{
  "error": "unauthorized"
}

# ACCEPTED: Correct X-API-KEY is given in HTTP header
curl -s -H "X-API-KEY: demo_key" http://localhost:8000/0 | jq
{
  "count": 1,
  "total": 1,
  "keys": [
    {
      "key": "key:1",
      "type": "rediseen"
    }
  ]
}
```
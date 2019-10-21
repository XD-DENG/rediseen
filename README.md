# rediseen

[![License](https://img.shields.io/:license-apache2-green.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![travis](https://api.travis-ci.org/XD-DENG/rediseen.svg?branch=master)](https://travis-ci.org/XD-DENG/rediseen/branches)
[![codecov](https://codecov.io/gh/XD-DENG/rediseen/branch/master/graph/badge.svg)](https://codecov.io/gh/XD-DENG/rediseen)


Start a REST-like API service for your Redis database, without writing a single line of code.


## 1. Quick Start

Let's assume that your Redis database URI is `redis://:@localhost:6379`, and you want to expose keys with prefix `key:` in logical database `0`.

```bash
# Install using Homebrew
brew install XD-DENG/rediseen/rediseen

# Configuration
export REDISEEN_REDIS_URI="redis://:@localhost:6379"
export REDISEEN_DB_EXPOSED=0
export REDISEEN_KEY_PATTERN_EXPOSED="^key:([0-9a-z]+)"

# Start the service
rediseen start
```

Now you should be able to query against the database, like `http://localhost:8000/0/key:1` (say you
have a key named `key:1` set in your Redis database).

For more details, please refer to the rest of this README documentation.



## 2. Usage


### 2.1 How to Install 

You can choose to install `Rediseen` either using `Homebrew` or from source.

- **Install Using `Homebrew`**

You can use [Homebrew](https://brew.sh/) to install `Rediseen`, no matte you are using `macOS`, or `Linux`/
`Windows 10 Subsystem for Linux` ([how to install Homebrew](https://docs.brew.sh/Installation)).

```bash
brew install XD-DENG/rediseen/rediseen

rediseen help
```

- **Build from source** (with Go 1.12 or above installed)

You can also build `Rediseen` from source.

```bash
git clone https://github.com/XD-DENG/rediseen.git
cd rediseen
go build . # executable binary file "rediseen" will be created

./rediseen help
```


### 2.2 How to Configure

Configuration is done via **environment variables**.

| Item | Description | Remark |
| --- | --- | --- |
| `REDISEEN_REDIS_URI` | URI of your Redis database, e.g. `redis://:@localhost:6379` | Compulsory |
| `REDISEEN_HOST` | Host of the service. Host will be `localhost` if `REDISEEN_HOST` is not explicitly set. Set to `0.0.0.0` if you want to expose your service to the world. | Optional |
| `REDISEEN_PORT` | Port of the service. Default port is 8000. | Optional |
| `REDISEEN_DB_EXPOSED` | Redis logical database(s) to expose.<br><br>E.g., `0`, `0;3;9`, `0-9;15`, or `*` (expose all logical databases) | Compulsory |
| `REDISEEN_KEY_PATTERN_EXPOSED` | Regular expression pattern, representing the name pattern of keys that you intend to expose.<br><br>For example, `user:([0-9a-z/.]+)\|^info:([0-9a-z/.]+)` exposes keys like `user:1`, `user:x1`, `testuser:1`, `info:1`, etc. |  |
| `REDISEEN_KEY_PATTERN_EXPOSE_ALL` | If you intend to expose ***all*** your keys, set `REDISEEN_KEY_PATTERN_EXPOSE_ALL` to `true`. | `REDISEEN_KEY_PATTERN_EXPOSED` can only be empty (or not set) if you have set `REDISEEN_KEY_PATTERN_EXPOSE_ALL` to `true`. |
| `REDISEEN_TEST_MODE` | Set to `true` to skip Redis connection validation for unit tests. | For Dev Only |


### 2.3 How to Start the Service

Run command below,

```bash
rediseen start
```

Then you can access the service at
- `http://<your server address>:<REDISEEN_PORT>/<redis DB>/<key>`
- `http://<your server address>:<REDISEEN_PORT>/<redis DB>/<key>/<index or value or member>`


### 2.4 How to Consume the Service

#### 2.4.1 `/<redis DB>/<key>`

| Data Type | Underlying Redis Command |
| --- | --- |
| STRING | `GET(key)` |
| LIST   | `LRANGE(key, 0, -1)` |
| SET    | `SMEMBERS(key)` |
| HASH   | `HGETALL(key)` |
| ZSET   | `ZRANGE(key, 0, -1)` |


#### 2.4.2 `/<redis DB>/<key>/<index or value or member>`

| Data Type | Usage | Return Value |
| --- | --- | --- |
| STRING | `/<redis DB>/<key>/<index>`  | `<index>`-th character in the string |
| LIST   | `/<redis DB>/<key>/<index>` | `<index>`-th element in the list |
| SET    | `/<redis DB>/<key>/<member>` | if `<member>` is member of the set |
| HASH   | `/<redis DB>/<key>/<field>` | value of hash `<field>` in the hash |
| ZSET   | `/<redis DB>/<key>/<memeber>` | index of `<member>` in the sorted set |



## 3. License

[Apache-2.0](https://www.apache.org/licenses/LICENSE-2.0)

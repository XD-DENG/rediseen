package conn

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis"
	"net/http"
	"os"
	"rediseen/types"
	"regexp"
	"strconv"
	"strings"
)

const strNotImplemented = "not implemented"
const strWrongTypeForIndexField = "wrong type for index/field"

// Client prepares a Redis client. Only Redis DB is needed, as all other information will be provided via configuration
func Client(db int) *redis.Client {
	parsedUri, _ := redis.ParseURL(os.Getenv("REDISEEN_REDIS_URI"))

	client := redis.NewClient(&redis.Options{
		Addr:     parsedUri.Addr,
		Password: parsedUri.Password,
		DB:       db,
	})

	return client
}

// ClientPing checks the user-specified `REDISEEN_REDIS_URI` (using default db 0)
func ClientPing() error {
	client := Client(0)
	defer client.Close()

	pingResult, err := client.Ping().Result()
	if pingResult != "PONG" {
		return err
	}

	return nil
}

// ListKeysByDb lists keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED,
// together with their typesGiven a Redis client (in which logical DB is specified),
// given a Redis client (in which logical DB is specified).
// Only up to 1000 keys will be returned.
// In the response, we also give `count` and `total`.
// `count`<=1000, while `total` is the actual total number of keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED
// Results are written into a http.ResponseWriter directly.
func ListKeysByDb(client *redis.Client, res http.ResponseWriter, regexpKeyPatternExposed *regexp.Regexp) {
	keys, _ := client.Keys("*").Result()

	var results []types.KeyInfoType

	for _, k := range keys {
		if regexpKeyPatternExposed.MatchString(k) {
			results = append(results, types.KeyInfoType{Key: k, Type: client.Type(k).Val()})
		}
	}

	var count int

	if len(results) >= 1000 {
		count = 1000
	} else {
		count = len(results)
	}

	js, _ := json.Marshal(types.KeyListType{Keys: results, Total: len(results), Count: count})
	res.Write(js)
}

// Get handles requests to different Redis Data Types, and return values correspondingly
func Get(client *redis.Client, res http.ResponseWriter, key string, indexOrField string) {

	var js []byte
	var index int64
	var field string
	var value interface{}

	keyType, err := client.Type(key).Result()

	if indexOrField == "" {
		switch keyType {
		case "string":
			value, err = client.Get(key).Result()
		case "list":
			value, err = client.LRange(key, 0, -1).Result()
		case "set":
			value, err = client.SMembers(key).Result()
		case "hash":
			value, err = client.HGetAll(key).Result()
		case "zset":
			//TODO: a simple implementation given methods on sorted set can be very complicated
			value, err = client.ZRange(key, 0, -1).Result()
		default:
			err = errors.New(strNotImplemented)
		}
	} else {
		if keyType == "string" || keyType == "list" {
			index, _ = strconv.ParseInt(indexOrField, 10, 64)
		} else {
			field = indexOrField
		}

		switch keyType {
		case "string":
			if index == 0 && indexOrField != "0" {
				err = errors.New(strWrongTypeForIndexField)
			} else {
				value, err = client.GetRange(key, index, index).Result()
			}
		case "list":
			if index == 0 && indexOrField != "0" {
				err = errors.New(strWrongTypeForIndexField)
			} else {
				value, err = client.LIndex(key, index).Result()
			}
		case "set":
			value, err = client.SIsMember(key, field).Result()
		case "hash":
			value, err = client.HGet(key, field).Result()
		case "zset":
			value, err = client.ZRank(key, field).Result()
		default:
			err = errors.New(strNotImplemented)
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), strNotImplemented) {
			res.WriteHeader(http.StatusNotImplemented)
		} else if strings.Contains(err.Error(), strWrongTypeForIndexField) {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.WriteHeader(http.StatusNotFound)
		}
		js, _ = json.Marshal(types.ErrorType{Error: err.Error()})
	} else {
		js, _ = json.Marshal(types.ResponseType{ValueType: keyType, Value: value})
	}
	res.Write(js)
}

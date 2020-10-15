package conn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/xd-deng/rediseen/types"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const strNotImplemented = "not implemented"
const strWrongTypeForIndexField = "wrong type for index/field"
const listKeyLimit = 1000

// ExtendedClient is a struct type which helps extend Redis Client
type ExtendedClient struct {
	RedisClient *redis.Client
}

var ctx = context.Background()

// Init initiates and prepares a Redis client.
// Only Redis DB is needed as argument, since all other information will be provided via configuration
func (client *ExtendedClient) Init(db int) {
	parsedUri, _ := redis.ParseURL(os.Getenv("REDISEEN_REDIS_URI"))

	client.RedisClient = redis.NewClient(&redis.Options{
		Addr:     parsedUri.Addr,
		Password: parsedUri.Password,
		DB:       db,
	})
}

// ClientPing checks the user-specified `REDISEEN_REDIS_URI` (using default db 0)
func ClientPing() error {

	var client ExtendedClient
	client.Init(0)
	defer client.RedisClient.Close()

	pingResult, err := client.RedisClient.Ping(ctx).Result()
	if pingResult != "PONG" {
		return err
	}

	return nil
}

// ListKeys lists keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED,
// together with their types,
// given a Redis client (in which logical DB is already specified).
// Only up to 1000 keys will be returned.
// In the response, we also give `count` and `total`.
// `count`<=1000, while `total` is the actual total number of keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED
func (client *ExtendedClient) ListKeys(regexpKeyPatternExposed *regexp.Regexp) []byte {
	keys, _ := client.RedisClient.Keys(ctx, "*").Result()

	pipe := client.RedisClient.Pipeline()

	var results []types.KeyInfoType
	for i, k := range keys {
		if i == listKeyLimit {
			break
		}
		if regexpKeyPatternExposed.MatchString(k) {
			pipe.Type(ctx, k)
			results = append(results, types.KeyInfoType{Key: k})
		}
	}

	pipeResult, _ := pipe.Exec(ctx)

	for i, cmder := range pipeResult {
		cmd := cmder.(*redis.StatusCmd)
		results[i].Type = cmd.Val()
	}

	var count int

	if len(results) >= listKeyLimit {
		count = listKeyLimit
	} else {
		count = len(results)
	}

	js, _ := json.Marshal(types.KeyListType{Keys: results, Total: len(keys), Count: count})
	return js
}

// Retrieve handles requests to different Redis Data Types, and return values correspondingly
func (client *ExtendedClient) Retrieve(key string, indexOrField string) ([]byte, int) {

	var js []byte
	var index int64
	var field string
	var value interface{}

	keyType, err := client.RedisClient.Type(ctx, key).Result()

	if indexOrField == "" {
		switch keyType {
		case "string":
			value, err = client.RedisClient.Get(ctx, key).Result()
		case "list":
			value, err = client.RedisClient.LRange(ctx, key, 0, -1).Result()
		case "set":
			value, err = client.RedisClient.SMembers(ctx, key).Result()
		case "hash":
			value, err = client.RedisClient.HGetAll(ctx, key).Result()
		case "zset":
			//TODO: a simple implementation given methods on sorted set can be very complicated
			value, err = client.RedisClient.ZRange(ctx, key, 0, -1).Result()
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
				value, err = client.RedisClient.GetRange(ctx, key, index, index).Result()
			}
		case "list":
			if index == 0 && indexOrField != "0" {
				err = errors.New(strWrongTypeForIndexField)
			} else {
				value, err = client.RedisClient.LIndex(ctx, key, index).Result()
			}
		case "set":
			value, err = client.RedisClient.SIsMember(ctx, key, field).Result()
		case "hash":
			value, err = client.RedisClient.HGet(ctx, key, field).Result()
		case "zset":
			value, err = client.RedisClient.ZRank(ctx, key, field).Result()
		default:
			err = errors.New(strNotImplemented)
		}
	}

	var errorCode int
	if err != nil {
		if strings.Contains(err.Error(), strNotImplemented) {
			errorCode = http.StatusNotImplemented
		} else if strings.Contains(err.Error(), strWrongTypeForIndexField) {
			errorCode = http.StatusBadRequest
		} else {
			errorCode = http.StatusNotFound
		}
		js, _ = json.Marshal(types.ErrorType{Error: err.Error()})
	} else {
		js, _ = json.Marshal(types.ResponseType{ValueType: keyType, Value: value})
	}

	return js, errorCode
}

// RedisInfo takes the results of Redis INFO command, then return the result as JSON ([]byte format from json.Marshal)
func (client *ExtendedClient) RedisInfo(section string, format string) ([]byte, error) {
	var infoResult string
	var err error
	if section == "" {
		section = "all"
	}
	infoResult, err = client.RedisClient.Info(ctx, section).Result()

	switch format {
	case "json":
		if infoResult == "" {
			return []byte{}, fmt.Errorf("invalid section `%s` is given. Check /info for supported sections", section)
		}
		if err != nil {
			return []byte{}, err
		}

		mapResult := make(map[string]map[string]string)
		var sectionName string
		for _, row := range strings.Split(infoResult, "\n") {
			if len(row) > 0 && string(row[0]) == "#" {
				// this row is the line for section name
				sectionName = strings.Trim(row, "\r# ")
				mapResult[sectionName] = make(map[string]string)
			} else {
				// this row is the line for detailed key-value pair
				values := strings.Split(row, ":")
				if len(values) != 2 {
					continue
				}
				mapResult[sectionName][values[0]] = strings.TrimSpace(values[1])
			}
		}

		jsonResult, _ := json.Marshal(mapResult)
		return jsonResult, nil
	case "prometheus":
		lines := strings.Split(infoResult, "\n")
		var filtered []string
		for _, l := range lines {
			processedLines := parseInfoLine(l)
			for _, pl := range processedLines {
				filtered = append(filtered, pl)
			}
		}
		return []byte(strings.ReplaceAll(strings.Join(filtered, "\n"), ":", " ")), nil
	case "raw":
		return []byte(infoResult), nil
	default:
		return []byte(infoResult), nil
	}
}

func validateFloatValue(v string) bool {
	_, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return true
	}
	return false
}

func parseInfoLine(v string) []string {
	result := []string{}

	v = strings.ReplaceAll(v, "\r", "")

	if strings.HasPrefix(v, "#") {
		// Include the comment lines
		result = append(result, v)
	} else {
		if strings.Contains(v, ",") {
			// For "commandstats" and "keyspace"
			// They need special parsing
			split := strings.Split(v, ":")

			for _, mv := range strings.Split(split[1], ",") {
				mv = strings.ReplaceAll(mv, "=", " ")
				result = append(result, split[0]+"_"+mv)
			}
		} else {
			// for other metric lines
			// The check here will help exclude all lines with non-float value,
			// including lines like "used_memory_human:846.21K"
			if strings.Contains(v, ":") && validateFloatValue(strings.Split(v, ":")[1]) {
				result = append(result, strings.ReplaceAll(v, ":", " "))
			}
		}
	}

	return result
}

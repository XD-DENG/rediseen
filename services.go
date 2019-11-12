package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"os"
	"rediseen/conn"
	"rediseen/types"
	"regexp"
	"strconv"
	"strings"
)

// parseKeyAndIndex helps parse strings like "key/3" in request like "/0/key/3" into "key" and "3"
// It should also be able to handle cases like "`key/1`/5" (i.e., slash is part of the key or index/field)
func parseKeyAndIndex(restPath string) (string, string) {
	var key string
	var index string

	countBacktick := strings.Count(restPath, "`")
	if countBacktick > 0 && countBacktick%2 == 0 {
		if restPath[0] == '`' && restPath[len(restPath)-1] == '`' {
			// Check case like /0/`key`/`index`
			bothBackTickPattern, _ := regexp.MatchString("`(?P<Key>.+)`/`(?P<Index>.+)`", restPath)
			if bothBackTickPattern {
				p := regexp.MustCompile("`(?P<Key>.+)`/`(?P<Index>.+)`")
				key = p.FindStringSubmatch(restPath)[1]
				index = p.FindStringSubmatch(restPath)[2]
			} else {
				key = restPath[1:(len(restPath) - 1)]
			}
		} else {
			p := regexp.MustCompile("`(?P<Key>.+)`/(?P<Index>.+)")
			key = p.FindStringSubmatch(restPath)[1]
			index = p.FindStringSubmatch(restPath)[2]
		}
	} else {
		if restPath[0] == '`' && restPath[len(restPath)-1] == '`' {
			key = restPath[1:(len(restPath) - 1)]
		} else {
			p := regexp.MustCompile("(?P<Key>.+)/(?P<Index>.+)")
			key = p.FindStringSubmatch(restPath)[1]
			index = p.FindStringSubmatch(restPath)[2]
		}
	}
	return key, index
}

func apiKeyMatch(req *http.Request) bool {
	if req.Header.Get("X-API-KEY") == os.Getenv("REDISEEN_API_KEY") {
		return true
	}
	return false
}

// Given a Redis client (in which logical DB is specified),
// List keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED, together with their types.
// Only up to 1000 keys will be returned.
// In the response, we also give `count` and `total`.
// `count`<=1000, while `total` is the actual total number of keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED
// Results are written into a http.ResponseWriter directly.
func listKeysByDb(client *redis.Client, res http.ResponseWriter, regexpKeyPatternExposed *regexp.Regexp) {
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

// Handle requests to different Redis Data Types, and return values correspondingly
func get(client *redis.Client, res http.ResponseWriter, key string, indexOrField string) {

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

func service(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	var js []byte

	var authEnforced bool
	authEnforced = os.Getenv("REDISEEN_API_KEY") != ""

	if authEnforced {
		if !apiKeyMatch(req) {
			res.WriteHeader(http.StatusUnauthorized)
			js, _ = json.Marshal(types.ErrorType{Error: "unauthorized"})
			res.Write(js)
			return
		}
	}

	if req.Method != "GET" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		js, _ = json.Marshal(types.ErrorType{Error: fmt.Sprintf("Method %s is not allowed", req.Method)})
		res.Write(js)
		return
	}

	// Process URL Path into detailed information, like DB and Key
	log.Printf("Request Path: '%s'\n", req.URL.Path)
	arguments := strings.Split(req.URL.Path, "/")
	countArguments := len(arguments)

	if strings.HasSuffix(req.URL.Path, "/") || countArguments < 2 || countArguments > 4 {
		res.WriteHeader(http.StatusBadRequest)
		js, _ = json.Marshal(types.ErrorType{Error: "Usage: /db, /db/key, /db/key/index, or /db/key/field"})
		res.Write(js)
		return
	}

	var rawDb string
	var key string
	var index string

	rawDb = arguments[1]

	db, err := strconv.Atoi(rawDb)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		js, _ = json.Marshal(types.ErrorType{Error: "Provide an integer for DB"})
		res.Write(js)
		return
	}

	client := conn.Client(db)
	defer client.Close()

	var config configuration
	config.loadFromEnv()
	if !config.dbCheck(db) {
		res.WriteHeader(http.StatusForbidden)
		js, _ = json.Marshal(types.ErrorType{Error: fmt.Sprintf("DB %d is not exposed", db)})
		res.Write(js)
		return
	}

	if countArguments == 2 {
		// request type-1: /db
		listKeysByDb(client, res, config.regexpKeyPatternExposed)
		return
	} else if countArguments == 3 {
		// request type-2: /db/key
		key = arguments[2]
	} else {
		// request type-3: /db/key/index, or /db/key/field
		key, index = parseKeyAndIndex(strings.Join(arguments[2:], "/"))
	}

	if !config.regexpKeyPatternExposed.MatchString(key) {
		res.WriteHeader(http.StatusForbidden)
		js, _ = json.Marshal(types.ErrorType{Error: "Key pattern is forbidden from access"})
		res.Write(js)
		return
	}

	// Check if key exists, meanwhile check Redis connection
	keyExists, err := client.Exists(key).Result()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		js, _ = json.Marshal(types.ErrorType{Error: err.Error()})
		res.Write(js)
		return
	}

	if keyExists == 0 {
		res.WriteHeader(http.StatusNotFound)
		js, _ = json.Marshal(types.ErrorType{Error: "Key provided does not exist."})
		res.Write(js)
		return
	}

	var logMsg strings.Builder
	logMsg.WriteString("Submit query for: db ")
	logMsg.WriteString(strconv.Itoa(db))
	logMsg.WriteString(", key `")
	logMsg.WriteString(key)
	logMsg.WriteString("`")
	if index != "" {
		logMsg.WriteString(", index/field `")
		logMsg.WriteString(index)
		logMsg.WriteString("`")
	}

	log.Println(logMsg.String())
	get(client, res, key, index)
}

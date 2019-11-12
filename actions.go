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

var dbExposedMap = make(map[int]bool)
var regexpKeyPatternAllowed *regexp.Regexp

// Check Configurations, and stop proceeding if any configuration is missing or conflicting
func configCheck() error {
	var redisURI = os.Getenv("REDISEEN_REDIS_URI")
	var dbExposed = os.Getenv("REDISEEN_DB_EXPOSED")
	var keyPatternAllowed = os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	var keyPatternAllowAll = os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")

	if os.Getenv("REDISEEN_API_KEY") == "" {
		log.Println("[WARNING] API is NOT secured with X-API-KEY")
	} else {
		log.Println("[INFO] API is secured with X-API-KEY (to access, specify X-API-KEY in request header)")
	}

	if redisURI == "" {
		return errors.New("No valid Redis URI is provided " +
			"(via environment variable REDISEEN_REDIS_URI)")
	}

	_, err := redis.ParseURL(redisURI)
	if err != nil {
		return fmt.Errorf("Redis URI provided "+
			"(via environment variable REDISEEN_REDIS_URI)"+
			"is not valid (details: %s)", err.Error())
	}

	if dbExposed == "" {
		return errors.New("REDISEEN_DB_EXPOSED is not configured")
	}

	errDbConfigCheckResult := validateDbExposeConfig(dbExposed)
	if errDbConfigCheckResult != nil {
		var errMsg strings.Builder
		errMsg.WriteString("REDISEEN_DB_EXPOSED provided can not be parsed properly")
		errMsg.WriteString(fmt.Sprintf(" (details: %s)", errDbConfigCheckResult.Error()))
		return errors.New(errMsg.String())
	}

	dbExposedMap = parseDbExposed(dbExposed)

	if keyPatternAllowAll == "true" {
		if keyPatternAllowed != "" {
			return errors.New("You have specified both REDISEEN_KEY_PATTERN_EXPOSED " +
				"and REDISEEN_KEY_PATTERN_EXPOSE_ALL=true, which is conflicting.")
		}
		log.Println("[WARNING] You are exposing ALL keys.")
	} else {
		if keyPatternAllowed == "" {
			strError := "You have not specified any key pattern to allow being accessed " +
				"(environment variable REDISEEN_KEY_PATTERN_EXPOSED)\n" +
				"        To allow ALL keys to be accessed, " +
				"set environment variable REDISEEN_KEY_PATTERN_EXPOSE_ALL=true"
			return errors.New(strError)
		} else {
			regexpKeyPatternAllowed, err = regexp.Compile(keyPatternAllowed)
			if err != nil {
				return fmt.Errorf("REDISEEN_KEY_PATTERN_EXPOSED can not be "+
					"compiled as regular expression. Details: %s\n", err.Error())
			}
			log.Println(fmt.Sprintf("[INFO] You are exposing keys of pattern `%s`", keyPatternAllowed))
		}
	}

	if os.Getenv("REDISEEN_TEST_MODE") != "true" {
		err = conn.ClientPing()
		if err != nil {
			return fmt.Errorf("Initial talking to Redis failed. "+
				"Please check the URI provided. Details: %s\n", err.Error())
		}
	}

	return nil
}

// validate if the string given as DB(s) to expose is legal.
// returns nil if it is legal, otherwise returns the error
func validateDbExposeConfig(configDbExposed string) error {
	// case-1: "*"
	if configDbExposed == "*" {
		log.Println("[WARNING] You are exposing ALL logical databases.")
		return nil
	}

	// case-2: "0" or "18"
	// case-3: "0-10" or "0;0-10" or "1-10;13"
	// If multiple values are provided (semicolon-separated), check value by value
	// This chunk handles cases where there is no semicolon in the string "automatically" as well
	parts := strings.Split(configDbExposed, ";")
	for _, p := range parts {
		subPatternCheck, _ := regexp.MatchString("(^[0-9]+$)|(^[0-9]+)(-)([0-9]+$)", p)

		if !subPatternCheck {
			return errors.New("illegal pattern")
		}
	}

	log.Println(fmt.Sprintf("[INFO] You are exposing logical database(s) `%s`", configDbExposed))
	return nil
}

// provide strings like "0;1;3;5" or "0;9-14;5" into a map for later querying
// store as a map to achieve O(1) search complexity
func parseDbExposed(configDbExposed string) map[int]bool {
	result := make(map[int]bool)
	parts := strings.Split(configDbExposed, ";")

	for _, p := range parts {
		patternCheckResult1, _ := regexp.MatchString("^[0-9]+$", p)
		patternCheckResult2, _ := regexp.MatchString("(^[0-9]+)(-)([0-9]+$)", p)

		if patternCheckResult1 {
			dbInt, _ := strconv.Atoi(p)
			result[dbInt] = true
		} else if patternCheckResult2 {
			temp := strings.Split(p, "-")
			dbInt1, _ := strconv.Atoi(temp[0])
			dbInt2, _ := strconv.Atoi(temp[1])
			for i := dbInt1; i <= dbInt2; i++ {
				result[i] = true
			}
		}
	}

	return result
}

//Check if db given by user is forbidden from being exposed
func dbCheck(db int, dbExposedMap map[int]bool) bool {
	if os.Getenv("REDISEEN_DB_EXPOSED") == "*" {
		return true
	}

	_, ok := dbExposedMap[db]
	if !ok {
		return false
	}
	return true
}

// Check if a string matches a pre-specified `keyPatternAllowed` (returns Boolean)
func keyPatternCheck(key string, regexpKeyPatternExposed *regexp.Regexp) bool {
	return regexpKeyPatternExposed.MatchString(key)
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
		if keyPatternCheck(k, regexpKeyPatternExposed) {
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

package main

import (
	"encoding/json"
	"fmt"
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
	config.validate()
	if !dbCheck(db, config.dbExposedMap) {
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

	if !keyPatternCheck(key, config.regexpKeyPatternExposed) {
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

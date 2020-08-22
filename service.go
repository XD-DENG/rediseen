package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/xd-deng/rediseen/conn"
	"github.com/xd-deng/rediseen/types"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type service struct {
	port                    string
	host                    string
	bindAddress             string
	redisURI                string
	dbExposed               string
	dbExposedMap            map[int]bool
	keyPatternExposed       string
	keyPatternExposeAll     bool
	apiKey                  string
	authEnforced            bool
	testMode                bool
	regexpKeyPatternExposed *regexp.Regexp
}

var ctx = context.Background()

func (c *service) loadConfigFromEnv() error {
	c.port = os.Getenv("REDISEEN_PORT")
	c.host = os.Getenv("REDISEEN_HOST")
	c.redisURI = os.Getenv("REDISEEN_REDIS_URI")
	c.dbExposed = os.Getenv("REDISEEN_DB_EXPOSED")
	c.keyPatternExposed = os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	c.keyPatternExposeAll = os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL") == "true"
	c.testMode = os.Getenv("REDISEEN_TEST_MODE") == "true"
	c.apiKey = os.Getenv("REDISEEN_API_KEY")

	if c.host == "" {
		c.host = defaultHost
	}
	if c.port == "" {
		c.port = defaultPort
	}
	c.bindAddress = net.JoinHostPort(c.host, c.port)

	if c.apiKey != "" {
		c.authEnforced = true
	}

	var err error

	if c.redisURI == "" {
		return errors.New("No valid Redis URI is provided (via environment variable REDISEEN_REDIS_URI)")
	}

	_, err = redis.ParseURL(c.redisURI)
	if err != nil {
		return fmt.Errorf("Redis URI provided (via environment variable REDISEEN_REDIS_URI)"+
			"is not valid (details: %s)", err.Error())
	}

	if c.dbExposed == "" {
		return errors.New("REDISEEN_DB_EXPOSED is not configured")
	}

	errDbConfigCheckResult := validateDbExposeConfig(c.dbExposed)
	if errDbConfigCheckResult != nil {
		var errMsg strings.Builder
		errMsg.WriteString("REDISEEN_DB_EXPOSED provided can not be parsed properly")
		errMsg.WriteString(fmt.Sprintf(" (details: %s)", errDbConfigCheckResult.Error()))
		return errors.New(errMsg.String())
	}

	c.dbExposedMap = parseDbExposed(c.dbExposed)

	if c.authEnforced {
		log.Println("[INFO] API is secured with X-API-KEY (to access, specify X-API-KEY in request header)")
	} else {
		log.Println("[WARNING] API is NOT secured with X-API-KEY")
	}

	c.regexpKeyPatternExposed, err = regexp.Compile(c.keyPatternExposed)
	if err != nil {
		return fmt.Errorf("REDISEEN_KEY_PATTERN_EXPOSED can not be "+
			"compiled as regular expression. Details: %s\n", err.Error())
	}

	if c.keyPatternExposeAll {
		if c.keyPatternExposed != "" {
			return errors.New("You have specified both REDISEEN_KEY_PATTERN_EXPOSED " +
				"and REDISEEN_KEY_PATTERN_EXPOSE_ALL=true, which is conflicting.")
		}
		log.Println("[WARNING] You are exposing ALL keys.")
	} else {
		if c.keyPatternExposed == "" {
			strError := "You have not specified any key pattern to allow being accessed " +
				"(environment variable REDISEEN_KEY_PATTERN_EXPOSED)\n" +
				"        To allow ALL keys to be accessed, " +
				"set environment variable REDISEEN_KEY_PATTERN_EXPOSE_ALL=true"
			return errors.New(strError)
		}
		log.Println(fmt.Sprintf("[INFO] You are exposing keys of pattern `%s`", c.keyPatternExposed))
	}

	if !c.testMode {
		err = conn.ClientPing()
		if err != nil {
			return fmt.Errorf("Initial talking to Redis failed. "+
				"Please check the URI provided. Details: %s\n", err.Error())
		}
	}
	return nil
}

//Check if db given by user is forbidden from being exposed
func (c *service) dbCheck(db int) bool {
	if c.dbExposed == "*" {
		return true
	}

	_, ok := c.dbExposedMap[db]
	if !ok {
		return false
	}
	return true
}

func (c *service) apiKeyMatch(req *http.Request) bool {
	if req.Header.Get("X-API-KEY") == c.apiKey {
		return true
	}
	return false
}

func (c *service) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	log.Printf("%s %s '%s' %s\n", req.Method, req.RemoteAddr, req.URL.Path, req.UserAgent())

	var js []byte

	if c.authEnforced {
		if !c.apiKeyMatch(req) {
			res.WriteHeader(http.StatusUnauthorized)
			js, _ = json.Marshal(types.ErrorType{Error: "unauthorized"})
			res.Write(js)
			log.Println("Unauthorized request")
			return
		}
	}

	if req.Method != "GET" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		js, _ = json.Marshal(types.ErrorType{Error: fmt.Sprintf("Method %s is not allowed", req.Method)})
		res.Write(js)
		log.Println("Method not allowed")
		return
	}

	// Process URL Path into detailed information, like DB and Key
	arguments := strings.Split(req.URL.Path, "/")
	countArguments := len(arguments)

	if req.URL.Path == "/" {
		res.Header().Set("Content-Type", "text/plain")
		res.Write([]byte(strHeader))
		res.Write([]byte("\n\n"))
		res.Write([]byte("Available Endpoints:\n - /info\n - /info/<info_section>\n - /<db>\n - /<db>/<key>\n - /<db>/<key>/<index>\n - /<db>/<key>/<field>"))
		return
	}

	if strings.HasSuffix(req.URL.Path, "/") || countArguments < 2 || countArguments > 4 {
		res.WriteHeader(http.StatusBadRequest)
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		encoder.Encode(types.ErrorType{Error: "Usage: /info, /info/<info_section>, /<db>, /<db>/<key>, /<db>/<key>/<index>, or /<db>/<key>/<field>"})
		res.Write(buffer.Bytes())
		return
	}

	var pathPart1 string // pathPart1 is either logical DB index OR "info"
	var pathPart2 string // pathPart2 is either key name in Redis database OR section name for Redis INFO command
	var pathPart3 string // pathPart3 is either index or field when clients submit queries like /<db>/<key>/<index> or /<db>/<key>/<field>

	pathPart1 = arguments[1]
	if countArguments == 3 {
		pathPart2 = arguments[2]
		// When clients query "/<DB>/<key>", <key> is case-sensitive
		// When clients query "/info/<section>", <section> is case-insensitive
		if pathPart1 == "info" {
			pathPart2 = strings.ToLower(pathPart2)
		}
	}
	if countArguments == 4 {
		pathPart2, pathPart3 = parseKeyAndIndex(strings.Join(arguments[2:], "/"))
	}

	if pathPart1 == "info" {

		var client conn.ExtendedClient
		client.Init(0)
		defer client.RedisClient.Close()

		js, err := client.RedisInfo(pathPart2)
		if err != nil {
			if strings.Contains(err.Error(), "invalid section") {
				res.WriteHeader(http.StatusBadRequest)
			} else {
				res.WriteHeader(http.StatusInternalServerError)
			}
			js, _ = json.Marshal(types.ErrorType{Error: "Exception while getting Redis Info. Details: " + err.Error()})
		}
		res.Write(js)
		return
	}

	db, err := strconv.Atoi(pathPart1)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		js, _ = json.Marshal(types.ErrorType{Error: "Provide an integer for DB"})
		res.Write(js)
		return
	}

	var client conn.ExtendedClient
	client.Init(db)
	defer client.RedisClient.Close()

	if !c.dbCheck(db) {
		res.WriteHeader(http.StatusForbidden)
		js, _ = json.Marshal(types.ErrorType{Error: fmt.Sprintf("DB %d is not exposed", db)})
		res.Write(js)
		return
	}

	if countArguments == 2 {
		// request type-1: /db
		res.Write(client.ListKeys(c.regexpKeyPatternExposed))
		return
	}

	if !c.regexpKeyPatternExposed.MatchString(pathPart2) {
		res.WriteHeader(http.StatusForbidden)
		js, _ = json.Marshal(types.ErrorType{Error: "Key pattern is forbidden from access"})
		res.Write(js)
		return
	}

	// Check if key exists, meanwhile check Redis connection
	keyExists, err := client.RedisClient.Exists(ctx, pathPart2).Result()
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
	logMsg.WriteString(pathPart2)
	logMsg.WriteString("`")
	if pathPart3 != "" {
		logMsg.WriteString(", index/field `")
		logMsg.WriteString(pathPart3)
		logMsg.WriteString("`")
	}

	log.Println(logMsg.String())
	js, errorCode := client.Retrieve(pathPart2, pathPart3)
	if errorCode != 0 {
		res.WriteHeader(errorCode)
	}
	res.Write(js)
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

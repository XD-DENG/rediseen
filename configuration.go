package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"rediseen/conn"
	"regexp"
	"strconv"
	"strings"
)

type configuration struct {
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

func (c *configuration) loadFromEnv() error {
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
	c.bindAddress = c.host + ":" + c.port

	if c.apiKey != "" {
		c.authEnforced = true
	}

	var err error
	c.regexpKeyPatternExposed, err = regexp.Compile(c.keyPatternExposed)
	if err != nil {
		return fmt.Errorf("REDISEEN_KEY_PATTERN_EXPOSED can not be "+
			"compiled as regular expression. Details: %s\n", err.Error())
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

	if c.redisURI == "" {
		return errors.New("No valid Redis URI is provided (via environment variable REDISEEN_REDIS_URI)")
	}

	_, err = redis.ParseURL(c.redisURI)
	if err != nil {
		return fmt.Errorf("Redis URI provided (via environment variable REDISEEN_REDIS_URI)"+
			"is not valid (details: %s)", err.Error())
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
func (c *configuration) dbCheck(db int) bool {
	if os.Getenv("REDISEEN_DB_EXPOSED") == "*" {
		return true
	}

	_, ok := c.dbExposedMap[db]
	if !ok {
		return false
	}
	return true
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

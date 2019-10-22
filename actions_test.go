package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_generateAddr(t *testing.T) {
	if generateAddr() != "localhost:8000" {
		t.Error("generateAddr is not handling default set-up correctly.")
	}
	
	os.Setenv("REDISEEN_HOST", "0.0.0.0")
	os.Setenv("REDISEEN_PORT", "80")
	generateAddr() != "0.0.0.0:80" {
		t.Error("generateAddr is not handling customized set-up correctly.")
	}
}

func Test_configCheck_no_redis_uri(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "No valid Redis URI is provided") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_invalid_redis_uri(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "mysql://a:b@localhost:8888/db")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "is not valid") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_no_db_exposed(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_DB_EXPOSED is not configured") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_invalid_db_exposed_1(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "-1")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_DB_EXPOSED provided can not be parsed properly") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_invalid_db_exposed_2(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "1;-2;10")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_DB_EXPOSED provided can not be parsed properly") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_no_key_pattern_specified(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	originalKeyPatternAllowAll := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", "")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", originalKeyPatternAllowAll)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "You have not specified any key pattern to allow being accessed") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}

	if !strings.Contains(err.Error(), "To allow ALL keys to be accessed,") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_conflicting_key_pattern_specified(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.]*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	originalKeyPatternAllowAll := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", "true")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", originalKeyPatternAllowAll)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "You have specified both") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}

	if !strings.Contains(err.Error(), "which is conflicting.") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_bad_regex(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_KEY_PATTERN_EXPOSED can not be compiled as regular expression") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_good_config(t *testing.T) {

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.]*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	err := configCheck()

	if err != nil {
		t.Error("Not expecting error but got error")
	}
}

func Test_configCheck_connection_failure(t *testing.T) {

	originalTestMode := os.Getenv("REDISEEN_TEST_MODE")
	os.Setenv("REDISEEN_TEST_MODE", "")
	defer os.Setenv("REDISEEN_TEST_MODE", originalTestMode)

	err := configCheck()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "Initial talking to Redis failed.") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_parseDbExposed_1(t *testing.T) {

	result := parseDbExposed("1;3;9;100")

	for _, i := range []int{1, 3, 9, 100} {
		_, ok := result[i]
		if !ok {
			t.Error("parsing wrongly")
		}
	}

	for _, i := range []int{2, 8, 99, 101} {
		_, ok := result[i]
		if ok {
			t.Error("parsing wrongly")
		}
	}
}

func Test_parseDbExposed_2(t *testing.T) {

	result := parseDbExposed("1;3-9;12-15;100")

	for _, i := range []int{1, 3, 4, 5, 6, 7, 8, 9, 12, 13, 14, 15, 100} {
		_, ok := result[i]
		if !ok {
			t.Error("parsing wrongly")
		}
	}

	for _, i := range []int{2, 10, 16, 99, 101} {
		_, ok := result[i]
		if ok {
			t.Error("parsing wrongly")
		}
	}
}

func Test_parseDbExposed_3(t *testing.T) {

	result := parseDbExposed("0")

	_, ok := result[0]
	if !ok {
		t.Error("parsing wrongly")
	}

	for _, i := range []int{2, 10, 16, 99, 101, 1, 3, 4, 5, 6, 7, 8, 9, 12, 13, 14, 15, 100} {
		_, ok := result[i]
		if ok {
			t.Error("parsing wrongly")
		}
	}
}

func Test_validateDbExposeConfig(t *testing.T) {
	var ok bool
	var err error

	ok, err = validateDbExposeConfig("*")
	if err != nil || ok == false {
		t.Error("checkDbExpose() failed for '*'")
	}

	ok, err = validateDbExposeConfig("8")
	if err != nil || ok == false {
		t.Error("checkDbExpose() failed for '8'")
	}

	ok, err = validateDbExposeConfig("1-10")
	if err != nil || ok == false {
		t.Error("checkDbExpose() failed for '1-10'")
	}

	ok, err = validateDbExposeConfig("1;3;8")
	if err != nil || ok == false {
		t.Error("checkDbExpose() failed for '1;3;8'")
	}

	ok, err = validateDbExposeConfig("1;3-7;18")
	if err != nil || ok == false {
		t.Error("checkDbExpose() failed for '1;3-7;18'")
	}

	ok, err = validateDbExposeConfig("-1;18")
	if err == nil || ok == true {
		t.Error("checkDbExpose() passed WRONGLY for '-1;18'")
	}

	ok, err = validateDbExposeConfig("a;18")
	if err == nil || ok == true {
		t.Error("checkDbExpose() passed WRONGLY for 'a;18'")
	}
}

func Test_dbCheck(t *testing.T) {
	// Test Environment Variable: REDISEEN_DB_EXPOSED=0-5

	for i := 0; i <= 5; i++ {
		if dbCheck(i) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}

	for _, i := range []int{6, 10, 8, 16, 99, 101} {
		if dbCheck(i) == true {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

func Test_dbCheck_expose_all(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "*")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16} {
		if dbCheck(i) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

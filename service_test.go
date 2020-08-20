package main

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/xd-deng/rediseen/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func Test_generateAddr(t *testing.T) {
	var testService service
	testService.loadConfigFromEnv()
	if testService.bindAddress != "localhost:8000" {
		t.Error("bindAddress is not created for default set-up correctly.")
	}

	os.Setenv("REDISEEN_HOST", "0.0.0.0")
	testService.loadConfigFromEnv()
	if testService.bindAddress != "0.0.0.0:8000" {
		t.Error("bindAddress is not created for customized set-up correctly.")
	}

	os.Setenv("REDISEEN_PORT", "80")
	testService.loadConfigFromEnv()
	if testService.bindAddress != "0.0.0.0:80" {
		t.Error("bindAddress is not created for customized set-up correctly.")
	}

	os.Unsetenv("REDISEEN_HOST")
	os.Unsetenv("REDISEEN_PORT")
	testService.loadConfigFromEnv()
	if testService.bindAddress != "localhost:8000" {
		t.Error("bindAddress is not created for default set-up correctly.")
	}
}

func Test_configCheck_no_redis_uri(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	err := testService.loadConfigFromEnv()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "No valid Redis URI is provided") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_invalid_redis_uri(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "mysql://a:b@localhost:8888/db")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	err := testService.loadConfigFromEnv()

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

	var testService service
	err := testService.loadConfigFromEnv()

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

	var testService service
	err := testService.loadConfigFromEnv()

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

	var testService service
	err := testService.loadConfigFromEnv()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_DB_EXPOSED provided can not be parsed properly") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_no_key_pattern_specified(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	originalKeyPatternAllowAll := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", "")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", originalKeyPatternAllowAll)

	var testService service
	err := testService.loadConfigFromEnv()

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

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.]*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	originalKeyPatternAllowAll := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", "true")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", originalKeyPatternAllowAll)

	var testService service
	err := testService.loadConfigFromEnv()

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

func Test_configCheck_all_key_exposed(t *testing.T) {

	originalValue1 := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", "true")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSE_ALL", originalValue1)

	originalValue2 := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalValue2)

	var testService service
	err := testService.loadConfigFromEnv()

	if err != nil {
		t.Error("Not expecting error but got error")
	}
}

func Test_configCheck_bad_regex(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	var testService service
	err := testService.loadConfigFromEnv()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "REDISEEN_KEY_PATTERN_EXPOSED can not be compiled as regular expression") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_configCheck_good_config_without_auth_config(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.]*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	var testService service
	err := testService.loadConfigFromEnv()

	if err != nil {
		t.Error("Not expecting error but got error")
	}
}

func Test_configCheck_good_config_with_auth_config(t *testing.T) {

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6379")
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	os.Setenv("REDISEEN_API_KEY", "RandomKey")
	defer os.Unsetenv("REDISEEN_API_KEY")

	originalKeyPatternAllowed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "^key:[.]*")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternAllowed)

	var testService service
	err := testService.loadConfigFromEnv()

	if err != nil {
		t.Error("Not expecting error but got error")
	}
}

func Test_configCheck_connection_failure(t *testing.T) {

	originalTestMode := os.Getenv("REDISEEN_TEST_MODE")
	os.Setenv("REDISEEN_TEST_MODE", "")
	defer os.Setenv("REDISEEN_TEST_MODE", originalTestMode)

	var testService service
	err := testService.loadConfigFromEnv()

	if err == nil {
		t.Error("Expecting error but got nil")
	}

	if !strings.Contains(err.Error(), "Initial talking to Redis failed.") {
		t.Error(fmt.Sprintf("Error contents `%s` is not what's expected", err.Error()))
	}
}

func Test_validateDbExposeConfig_valid_cases(t *testing.T) {
	var err error

	testStrings := []string{"*", "8", "1-10", "1;3;8", "1;3-7;18", "1;3-7;10-100"}

	for _, s := range testStrings {
		err = validateDbExposeConfig(s)
		if err != nil {
			t.Error(fmt.Sprintf("checkDbExpose() failed for '%s'", s))
		}
	}
}

func Test_validateDbExposeConfig_invalid_cases(t *testing.T) {
	var err error

	testStrings := []string{"-1;18", "a;18", "1,18", ";1;18", "1;18;", "1;-5", "1;-5;10"}

	for _, s := range testStrings {
		err = validateDbExposeConfig(s)
		if err == nil {
			t.Error(fmt.Sprintf("checkDbExpose() passed WRONGLY for '%s'", s))
		}
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

	for i := 1; i <= 100; i++ {
		_, ok := result[i]
		if ok {
			t.Error("parsing wrongly")
		}
	}
}

func Test_dbCheck(t *testing.T) {
	// Test Environment Variable: REDISEEN_DB_EXPOSED=0-5

	var testService service
	testService.loadConfigFromEnv()
	for i := 0; i <= 5; i++ {
		if testService.dbCheck(i) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}

	for _, i := range []int{6, 10, 8, 16, 99, 101} {
		if testService.dbCheck(i) == true {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

func Test_dbCheck_expose_all(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "*")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	var testService service
	testService.loadConfigFromEnv()
	for i := 0; i <= 100; i++ {
		if testService.dbCheck(i) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

func compareAndShout(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Error("Expecting\n", expected, "\ngot\n", actual)
	}
}

func Test_service_wrong_usage(t *testing.T) {

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	expectedCode := 400
	expectedError := "Usage: /info, /info/<info_section>, /<db>, /<db>/<key>, /<db>/<key>/<index>, or /<db>/<key>/<field>"
	casesToTest := []string{"/0/", "/0/key:1/", "/0/key:1/1/", "/0/key:1/1/1", "/0/key:1/1/1/", "/0/key:1/1/1/1"}
	var res *http.Response

	for _, suffix := range casesToTest {
		res, _ = http.Get(s.URL + suffix)

		if res.StatusCode != expectedCode {
			t.Error("Expecting\n", expectedCode, "\ngot\n", res.StatusCode)
		}

		resultStr, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var result types.ErrorType
		json.Unmarshal([]byte(resultStr), &result)

		compareAndShout(t, expectedError, result.Error)
	}
}

func Test_service_non_integer_db_provided(t *testing.T) {

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/a/key")

	expectedCode := 400
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Provide an integer for DB"
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_redis_conn_refused(t *testing.T) {

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 500
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "connection refused"
	if !strings.Contains(result.Error, expectedError) {
		t.Error("Expecting to contain \n", expectedError, "\ngot\n", result.Error)
	}
}

func Test_service_non_existent_key(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 404
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Key provided does not exist."
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_validate_case_sensitive_1(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("KEY:1", "fake content")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 404
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Key provided does not exist."
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_validate_case_sensitive_2(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key", "fake content")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	originalKeyPatternExposed := os.Getenv("REDISEEN_KEY_PATTERN_EXPOSED")
	os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", "KEY")
	defer os.Setenv("REDISEEN_KEY_PATTERN_EXPOSED", originalKeyPatternExposed)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/KEY")

	expectedCode := 404
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Key provided does not exist."
	compareAndShout(t, expectedError, result.Error)
}

// Check listing-keys feature
// Taking keys which are NOT exposed into consideration as well (they should NOT be counted or returned)
func Test_service_list_keys_by_db_1(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	testSlice1 := []string{"hi", "world", "test", "done"}
	for i, v := range testSlice1 {
		mr.Set(fmt.Sprintf("key:%v", i), v)
	}

	testSlice2 := []string{"r", "a", "n", "d", "o", "m"}
	for i, v := range testSlice2 {
		mr.HSet(fmt.Sprintf("key:%v", i+len(testSlice1)), v, v)
	}

	testSlice3 := []string{"list", "type"}
	for i, v := range testSlice3 {
		mr.Lpush(fmt.Sprintf("key:%v", i+len(testSlice1)+len(testSlice2)), v)
	}

	testSlice4 := []string{"no", "access", "key"}
	for i, v := range testSlice4 {
		mr.HSet(fmt.Sprintf("no_access_key:%v", i+len(testSlice1)+len(testSlice2)+len(testSlice3)), v, v)
	}

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.KeyListType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, len(testSlice1)+len(testSlice2)+len(testSlice3), len(result.Keys))
	compareAndShout(t, len(testSlice1)+len(testSlice2)+len(testSlice3), result.Count)
	if result.Total > 1000 {
		t.Error("Listing keys function should return <= 1000 keys")
	}
}

// Check listing-keys feature
// Check the situation where more than 1000 keys are exposed
// `Total` should be the total number, BUT only up to 1000 keys should be returned.
// `Count` should be up to 1000 as well
func Test_service_list_keys_by_db_2(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	n := 2000
	for i := 0; i < 2000; i++ {
		mr.Set(fmt.Sprintf("key:%v", i), string(i))
	}

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.KeyListType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, 1000, len(result.Keys))
	for i := 0; i < 1000; i++ {
		compareAndShout(t, "string", result.Keys[i].Type)
	}
	compareAndShout(t, 1000, result.Count)
	compareAndShout(t, n, result.Total)
}

func Test_service_list_keys_by_db_key_type_list(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	n := 2000
	for i := 0; i < n; i++ {
		mr.Lpush(fmt.Sprintf("key:%v", i), string(i))
	}

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.KeyListType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, 1000, len(result.Keys))
	for i := 0; i < 1000; i++ {
		compareAndShout(t, "list", result.Keys[i].Type)
	}
	compareAndShout(t, 1000, result.Count)
	compareAndShout(t, n, result.Total)
}

func Test_service_list_keys_by_db_key_type_hash(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	n := 500
	for i := 0; i < n; i++ {
		mr.HSet(fmt.Sprintf("key:%v", i), string(i), string(i))
	}

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.KeyListType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, n, len(result.Keys))
	for i := 0; i < n; i++ {
		compareAndShout(t, "hash", result.Keys[i].Type)
	}
	compareAndShout(t, n, result.Count)
	compareAndShout(t, n, result.Total)
}

func Test_service_list_keys_by_db_key_type_mixed(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:string", "string")
	mr.HSet("key:hash", "k", "v")
	mr.Lpush("key:list", "element")
	mr.SetAdd("key:set", "hi")
	mr.ZAdd("key:zset", 5, "hi")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.KeyListType
	json.Unmarshal([]byte(resultStr), &result)

	for _, k := range result.Keys {
		switch k.Key {
		case "key_string":
			compareAndShout(t, "string", k.Type)
		case "key_hash":
			compareAndShout(t, "hash", k.Type)
		case "key_list":
			compareAndShout(t, "list", k.Type)
		case "key_set":
			compareAndShout(t, "set", k.Type)
		case "key_zset":
			compareAndShout(t, "zset", k.Type)
		}

	}
	compareAndShout(t, 5, len(result.Keys))
	compareAndShout(t, 5, result.Count)
	compareAndShout(t, 5, result.Total)
}

func Test_service_string_type(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "hi")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ResponseType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "hi", result.Value)
}

func Test_service_string_check_by_index(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "Developer")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ResponseType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "D", result.Value)

	res, _ = http.Get(s.URL + "/0/key:1/4")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "l", result.Value)

	res, _ = http.Get(s.URL + "/0/`key:1`/5")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "o", result.Value)

	res, _ = http.Get(s.URL + "/0/`key:1`/`6`")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "p", result.Value)
}

func Test_service_string_check_by_index_wrong_index(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "Developer")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/x")

	compareAndShout(t, 400, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "wrong type for index/field"
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_string_type_with_slash_in_key(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:/1", "hi")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/`key:/1`")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ResponseType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "hi", result.Value)
}

func Test_service_string_type_with_slash_and_backtick_in_key(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:`/1", "hi")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/`key:`/1`")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ResponseType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "string", result.ValueType)
	compareAndShout(t, "hi", result.Value)
}

// Validate if listing-keys feature returns 200 if the DB in client request IS exposed
func Test_service_list_keys_for_db_with_access(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	// env var set for the test is REDISEEN_DB_EXPOSED=0-5
	for db := 0; db <= 5; db++ {
		res, _ := http.Get(s.URL + fmt.Sprintf("/%v", db))

		expectedCode := 200
		compareAndShout(t, expectedCode, res.StatusCode)

		resultStr, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var result types.KeyListType
		json.Unmarshal([]byte(resultStr), &result)

		compareAndShout(t, 0, result.Count)
		compareAndShout(t, 0, result.Total)
	}
}

// Validate if listing-keys feature returns 403 if the DB in client request is NOT exposed
func Test_service_list_keys_for_db_without_access(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	// env var set for the test is REDISEEN_DB_EXPOSED=0-5
	for db := 6; db <= 100; db++ {
		res, _ := http.Get(s.URL + fmt.Sprintf("/%v", db))

		expectedCode := 403
		compareAndShout(t, expectedCode, res.StatusCode)

		resultStr, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var result types.ErrorType
		json.Unmarshal([]byte(resultStr), &result)

		expectedError := fmt.Sprintf("DB %v is not exposed", db)
		compareAndShout(t, expectedError, result.Error)
	}
}

func Test_service_string_type_db_no_access(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	// env var set for the test is REDISEEN_DB_EXPOSED=0-5
	for _, db := range []int{6, 10, 100} {
		res, _ := http.Get(s.URL + fmt.Sprintf("/%v/key:1", db))

		expectedCode := 403
		compareAndShout(t, expectedCode, res.StatusCode)

		resultStr, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var result types.ErrorType
		json.Unmarshal([]byte(resultStr), &result)

		expectedError := fmt.Sprintf("DB %v is not exposed", db)
		compareAndShout(t, expectedError, result.Error)
	}
}

func Test_service_string_type_key_no_access(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	//env var set for the test is REDISEEN_KEY_PATTERN_EXPOSED=^key:[.]*
	mr.Set("id:1", "hi")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/id:1")

	expectedCode := 403
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Key pattern is forbidden from access"
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_list_key(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Lpush("key:1", "hello")
	mr.Lpush("key:1", "world")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"list","value":["world","hello"]}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_list_key_check_by_index(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Lpush("key:1", "hello")
	mr.Lpush("key:1", "world")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/0")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"list","value":"world"}`
	compareAndShout(t, expectedResult, string(result))

	res, _ = http.Get(s.URL + "/0/key:1/1")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult = `{"type":"list","value":"hello"}`
	compareAndShout(t, expectedResult, string(result))

	// Check wrong type for index
	res, _ = http.Get(s.URL + "/0/key:1/a")

	expectedCode = 400
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult = fmt.Sprintf(`{"error":"%s"}`, "wrong type for index/field")
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_set(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.SetAdd("key:1", "hello")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"set","value":["hello"]}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_set_check_by_index(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.SetAdd("key:1", "hello")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/hello")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"set","value":true}`
	compareAndShout(t, expectedResult, string(result))

	res, _ = http.Get(s.URL + "/0/key:1/world")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult = `{"type":"set","value":false}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_hash(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.HSet("key:1", "role", "developer")
	mr.HSet("key:1", "id", "1")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"hash","value":{"id":"1","role":"developer"}}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_hash_check_by_index(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.HSet("key:1", "role", "developer")
	mr.HSet("key:1", "id", "1")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/role")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"hash","value":"developer"}`
	compareAndShout(t, expectedResult, string(result))

	res, _ = http.Get(s.URL + "/0/key:1/id")

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult = `{"type":"hash","value":"1"}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_zset(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.ZAdd("key:set", 100, "developer")
	mr.ZAdd("key:set", 0, "bluffer")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:set")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"zset","value":["bluffer","developer"]}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_zset_check_by_field(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.ZAdd("key:set", 200, "Mr.X")
	mr.ZAdd("key:set", 100, "developer")
	mr.ZAdd("key:set", 0, "bluffer")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:set/developer")

	expectedCode := 200
	compareAndShout(t, expectedCode, res.StatusCode)

	result, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	expectedResult := `{"type":"zset","value":1}`
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_delete_not_allowed(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "hello")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", s.URL+"/0/key:1", nil)
	res, _ := client.Do(req)

	expectedCode := 405
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Method DELETE is not allowed"
	compareAndShout(t, expectedError, result.Error)
}

// Request method will happen first, so whatever method other than GET should always get rejected
// Checks like key pattern check should not even happen
func Test_service_delete_not_allowed_no_access(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("id:1", "hello")

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", s.URL+"/0/id:1", nil)
	res, _ := client.Do(req)

	expectedCode := 405
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Method DELETE is not allowed"
	compareAndShout(t, expectedError, result.Error)
}

func Test_api_key_authentication(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "hello")

	os.Setenv("REDISEEN_API_KEY", "nopass")
	defer os.Setenv("REDISEEN_API_KEY", "")

	var testService service
	testService.loadConfigFromEnv()
	s := httptest.NewServer(http.Handler(&testService))
	defer s.Close()

	// case-1: no API Key is provided in request header
	// (to FAIL)
	res, _ := http.Get(s.URL + "/0")

	expectedCode := 401
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	compareAndShout(t, "unauthorized", result.Error)

	// case-2: correct API Key is provided in request header
	// (to SUCCEED)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", s.URL+"/0", nil)
	req.Header.Add("X-API-KEY", "nopass")
	res, _ = client.Do(req)

	expectedCode = 200
	compareAndShout(t, expectedCode, res.StatusCode)

	// case-3: wrong API Key is provided in request header
	// (to FAIL)
	client = &http.Client{}
	req, _ = http.NewRequest("GET", s.URL+"/0", nil)
	req.Header.Add("X-API-KEY", "wrongkey")
	res, _ = client.Do(req)

	expectedCode = 401
	compareAndShout(t, expectedCode, res.StatusCode)
}

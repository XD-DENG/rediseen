package main

import (
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"rediseen/types"
	"strings"
	"testing"
)

func compareAndShout(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Error("Expecting\n", expected, "\ngot\n", actual)
	}
}

func Test_service_wrong_usage(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(service))
	defer s.Close()

	expectedCode := 400
	expectedError := "Usage: /db/key or /db/key/<index or filed>"
	var res *http.Response

	for _, suffix := range []string{"/0", "/0/", "/0/key:1/", "/0/key:1/1/"} {
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

	s := httptest.NewServer(http.HandlerFunc(service))
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

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s1 := httptest.NewServer(http.HandlerFunc(service))
	defer s1.Close()

	res, _ := http.Get(s1.URL + "/0/key:1")

	expectedCode := 404
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "Key provided does not exist."
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_string_type(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:1", "hi")

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
	defer s.Close()

	res, _ := http.Get(s.URL + "/0/key:1/x")

	compareAndShout(t, 400, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := strWrongTypeForIndexField
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_string_type_with_slash_in_key(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.Set("key:/1", "hi")

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

func Test_service_string_type_db_no_access(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
	defer s.Close()

	// env var set for the test is REDISEEN_DB_EXPOSED=0-5
	res, _ := http.Get(s.URL + "/10/key:1")

	expectedCode := 403
	compareAndShout(t, expectedCode, res.StatusCode)

	resultStr, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var result types.ErrorType
	json.Unmarshal([]byte(resultStr), &result)

	expectedError := "DB 10 is not exposed"
	compareAndShout(t, expectedError, result.Error)
}

func Test_service_string_type_key_no_access(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	//env var set for the test is REDISEEN_KEY_PATTERN_EXPOSED=^key:[.]*
	mr.Set("id:1", "hi")

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	expectedResult = fmt.Sprintf(`{"error":"%s"}`, strWrongTypeForIndexField)
	compareAndShout(t, expectedResult, string(result))
}

func Test_service_set(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	mr.SetAdd("key:1", "hello")

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

	originalRedisUri := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisUri)

	s := httptest.NewServer(http.HandlerFunc(service))
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

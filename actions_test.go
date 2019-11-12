package main

import (
	"os"
	"testing"
)

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

	var config configuration
	config.loadFromEnv()
	for i := 0; i <= 5; i++ {
		if dbCheck(i, config.dbExposedMap) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}

	for _, i := range []int{6, 10, 8, 16, 99, 101} {
		if dbCheck(i, config.dbExposedMap) == true {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

func Test_dbCheck_expose_all(t *testing.T) {

	originalDbExposed := os.Getenv("REDISEEN_DB_EXPOSED")
	os.Setenv("REDISEEN_DB_EXPOSED", "*")
	defer os.Setenv("REDISEEN_DB_EXPOSED", originalDbExposed)

	for i := 0; i <= 100; i++ {
		if dbCheck(i, config.dbExposedMap) == false {
			t.Error("something is wrong with dbCheck()")
		}
	}
}

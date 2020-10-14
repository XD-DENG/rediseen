package conn

import (
	"fmt"
	"github.com/alicebob/miniredis"
	"os"
	"reflect"
	"testing"
)

func Test_ClientPing(t *testing.T) {

	mr, _ := miniredis.Run()
	defer mr.Close()

	originalRedisURI := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", fmt.Sprintf("redis://:@%s", mr.Addr()))
	defer os.Setenv("REDISEEN_REDIS_URI", originalRedisURI)

	err := ClientPing()
	if err != nil {
		t.Error("Not expecting error but got error")
	}
}

func Test_parseInfoLine(t *testing.T) {
	data := []string{
		"used_memory:866520",
		"redis_version:6.0.8",
		"# CPU",
		"used_memory_human:846.21K",
		"db0:keys=1053522,expires=7,avg_ttl=257671496273",
		"cmdstat_info:calls=2,usec=254,usec_per_call=127.00",
		"executable:/data/redis-server",
	}
	expected := [][]string{
		{"used_memory 866520"},
		{},
		{"# CPU"},
		{},
		{"db0_keys 1053522", "db0_expires 7", "db0_avg_ttl 257671496273"},
		{"cmdstat_info_calls 2", "cmdstat_info_usec 254", "cmdstat_info_usec_per_call 127.00"},
		{},
	}

	for i, x := range data {
		if !reflect.DeepEqual(parseInfoLine(x), expected[i]) {
			t.Error("Something is wrong")
		}
	}
}

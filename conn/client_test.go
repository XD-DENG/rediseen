package conn

import (
	"fmt"
	"github.com/alicebob/miniredis"
	"os"
	"testing"
)

func Test_service_non_existent_key(t *testing.T) {

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

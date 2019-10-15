package conn

import (
	"github.com/go-redis/redis"
	"os"
)

// Prepare a Redis client
// Only Redis DB is needed, as all other information will be provided via configuration
func Client(db int) *redis.Client {
	parsedUri, _ := redis.ParseURL(os.Getenv("REDISEEN_REDIS_URI"))

	client := redis.NewClient(&redis.Options{
		Addr:     parsedUri.Addr,
		Password: parsedUri.Password,
		DB:       db,
	})

	return client
}

// Check the user-specified `REDISEEN_REDIS_URI` (using default db 0)
func ClientPing() error {
	client := Client(0)
	defer client.Close()

	if os.Getenv("REDISEEN_TEST_MODE") != "true" {
		pingResult, err := client.Ping().Result()
		if pingResult != "PONG" {
			return err
		}
	}
	return nil
}

package conn

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis"
	"net/http"
	"os"
	"rediseen/types"
	"regexp"
	"strconv"
	"strings"
)

const strNotImplemented = "not implemented"
const strWrongTypeForIndexField = "wrong type for index/field"

type ExtendedClient struct {
	RedisClient *redis.Client
}

// Client prepares a Redis client. Only Redis DB is needed, as all other information will be provided via configuration
func (client *ExtendedClient) Init(db int) {
	parsedUri, _ := redis.ParseURL(os.Getenv("REDISEEN_REDIS_URI"))

	client.RedisClient = redis.NewClient(&redis.Options{
		Addr:     parsedUri.Addr,
		Password: parsedUri.Password,
		DB:       db,
	})
}

// ClientPing checks the user-specified `REDISEEN_REDIS_URI` (using default db 0)
func ClientPing() error {

	var client ExtendedClient
	client.Init(0)
	defer client.RedisClient.Close()

	pingResult, err := client.RedisClient.Ping().Result()
	if pingResult != "PONG" {
		return err
	}

	return nil
}

// ListKeys lists keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED,
// together with their types,
// given a Redis client (in which logical DB is already specified).
// Only up to 1000 keys will be returned.
// In the response, we also give `count` and `total`.
// `count`<=1000, while `total` is the actual total number of keys whose names match with REDISEEN_KEY_PATTERN_EXPOSED
// Results are written into a http.ResponseWriter directly.
func (client *ExtendedClient) ListKeys(res http.ResponseWriter, regexpKeyPatternExposed *regexp.Regexp) {
	keys, _ := client.RedisClient.Keys("*").Result()

	var results []types.KeyInfoType

	for _, k := range keys {
		if regexpKeyPatternExposed.MatchString(k) {
			results = append(results, types.KeyInfoType{Key: k, Type: client.RedisClient.Type(k).Val()})
		}
	}

	var count int

	if len(results) >= 1000 {
		count = 1000
	} else {
		count = len(results)
	}

	js, _ := json.Marshal(types.KeyListType{Keys: results, Total: len(results), Count: count})
	res.Write(js)
}

// Retrieve handles requests to different Redis Data Types, and return values correspondingly
func (client *ExtendedClient) Retrieve(res http.ResponseWriter, key string, indexOrField string) {

	var js []byte
	var index int64
	var field string
	var value interface{}

	keyType, err := client.RedisClient.Type(key).Result()

	if indexOrField == "" {
		switch keyType {
		case "string":
			value, err = client.RedisClient.Get(key).Result()
		case "list":
			value, err = client.RedisClient.LRange(key, 0, -1).Result()
		case "set":
			value, err = client.RedisClient.SMembers(key).Result()
		case "hash":
			value, err = client.RedisClient.HGetAll(key).Result()
		case "zset":
			//TODO: a simple implementation given methods on sorted set can be very complicated
			value, err = client.RedisClient.ZRange(key, 0, -1).Result()
		default:
			err = errors.New(strNotImplemented)
		}
	} else {
		if keyType == "string" || keyType == "list" {
			index, _ = strconv.ParseInt(indexOrField, 10, 64)
		} else {
			field = indexOrField
		}

		switch keyType {
		case "string":
			if index == 0 && indexOrField != "0" {
				err = errors.New(strWrongTypeForIndexField)
			} else {
				value, err = client.RedisClient.GetRange(key, index, index).Result()
			}
		case "list":
			if index == 0 && indexOrField != "0" {
				err = errors.New(strWrongTypeForIndexField)
			} else {
				value, err = client.RedisClient.LIndex(key, index).Result()
			}
		case "set":
			value, err = client.RedisClient.SIsMember(key, field).Result()
		case "hash":
			value, err = client.RedisClient.HGet(key, field).Result()
		case "zset":
			value, err = client.RedisClient.ZRank(key, field).Result()
		default:
			err = errors.New(strNotImplemented)
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), strNotImplemented) {
			res.WriteHeader(http.StatusNotImplemented)
		} else if strings.Contains(err.Error(), strWrongTypeForIndexField) {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.WriteHeader(http.StatusNotFound)
		}
		js, _ = json.Marshal(types.ErrorType{Error: err.Error()})
	} else {
		js, _ = json.Marshal(types.ResponseType{ValueType: keyType, Value: value})
	}
	res.Write(js)
}

func (client *ExtendedClient) RedisInfo(section string) (types.Info, error) {
	var infoResult string
	var err error
	if section == "" {
		infoResult, err = client.RedisClient.Info().Result()
	} else {
		infoResult, err = client.RedisClient.Info(section).Result()
	}
	if err != nil {
		return types.Info{}, err
	}

	mapResult := make(map[string]string)
	for _, kv := range strings.Split(infoResult, "\n") {
		values := strings.Split(kv, ":")
		if len(values) != 2 {
			continue
		}
		mapResult[values[0]] = strings.TrimSpace(values[1])
	}

	var result types.Info

	result.Server = types.InfoServer{
		RedisVersion:    mapResult["redis_version"],
		RedisBuildId:    mapResult["redis_build_id"],
		RedisMode:       mapResult["redis_mode"],
		Os:              mapResult["os"],
		ArchBits:        mapResult["arch_bits"],
		GccVersion:      mapResult["gcc_version"],
		ProcessId:       mapResult["process_id"],
		RunId:           mapResult["run_id"],
		TcpPort:         mapResult["tcp_port"],
		UptimeInSeconds: mapResult["uptime_in_seconds"],
		UptimeInDays:    mapResult["uptime_in_days"],
		Hz:              mapResult["hz"],
		ConfiguredHz:    mapResult["configured_hz"],
		LruClock:        mapResult["lru_clock"],
		Executable:      mapResult["executable"],
		ConfigFile:      mapResult["config_file"],
	}

	result.Clients = types.InfoClients{
		ConnectedClients: mapResult["connected_clients"],
		BlockedClients:   mapResult["blocked_clients"],
	}

	result.Replication = types.InfoReplication{
		Role:            mapResult["role"],
		ConnectedSlaves: mapResult["connected_slaves"],
		MasterReplId:    mapResult["master_replid"],
		MasterReplId2:   mapResult["master_replid2"],
	}

	result.Cpu = types.InfoCpu{
		UsedCpuSys:          mapResult["used_cpu_sys"],
		UsedCpuUser:         mapResult["used_cpu_user"],
		UsedCpuSysChildren:  mapResult["used_cpu_sys_children"],
		UsedCpuUserChildren: mapResult["used_cpu_user_children"],
	}

	result.Cluster = types.InfoCluster{ClusterEnabled: mapResult["cluster_enabled"]}

	return result, nil
}

package config

import (
	"log"
	"os"
	"strconv"
)

var (
	// export
	ApiKey      string
	NapcatWSURL string
	MasterID    uint64
	BotID       uint64

	PsqlHost     string
	PsqlPort     uint16
	PsqlUser     string
	PsqlPassword string
	PsqlDbName   string

	ExchangeRateAPIKey string
	OkxMirrorAPIKey    string
)

const (
	// environment values
	env_NAPCAT_HOST  = "NAPCAT_HOST"
	env_ACCESS_TOKEN = "ACCESS_TOKEN"
	env_API_KEY      = "API_KEY"
	env_MASTER_ID    = "MASTER_ID"
	env_BOT_ID       = "BOT_ID"

	env_PSQL_HOST     = "PSQL_HOST"
	env_PSQL_PORT     = "PSQL_PORT"
	env_PSQL_USER     = "PSQL_USER"
	env_PSQL_PASSWORD = "PSQL_PASSWORD"
	env_PSQL_DBNAME   = "PSQL_DBNAME"

	env_EXCHANGE_RATE_API_KEY = "EXCHANGE_RATE_API_KEY"
	env_OKX_MIRROR_API_KEY    = "OKX_MIRROR_API_KEY"
)

var BotOwnerGroupIDs = []uint64{
	948697448,
	866738031,
}

func getEnvString(env string, def string) string {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	return val
}

func getEnvUInt(env string, def uint64) uint64 {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	ret, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		log.Fatalf("Parse %s=%s failed: %s", env, val, err.Error())
	}
	return ret
}

func getEnvPort(env string, def uint16) uint16 {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	ret, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		log.Fatalf("Parse port %s=%s failed: %s", env, val, err.Error())
	}
	return uint16(ret)
}

func init() {
	napcatHost := getEnvString(env_NAPCAT_HOST, "127.0.0.1:3001")
	accessToken := os.Getenv(env_ACCESS_TOKEN)
	ApiKey = os.Getenv(env_API_KEY)

	NapcatWSURL = "ws://" + napcatHost
	if accessToken != "" {
		NapcatWSURL += "?access_token=" + accessToken
	}

	MasterID = getEnvUInt(env_MASTER_ID, 1006554341)
	BotID = getEnvUInt(env_BOT_ID, 3552586437)
	PsqlHost = getEnvString(env_PSQL_HOST, "127.0.0.1")
	PsqlPort = getEnvPort(env_PSQL_PORT, 5432)
	PsqlUser = os.Getenv(env_PSQL_USER)
	PsqlPassword = os.Getenv(env_PSQL_PASSWORD)
	PsqlDbName = os.Getenv(env_PSQL_DBNAME)
	ExchangeRateAPIKey = os.Getenv(env_EXCHANGE_RATE_API_KEY)
	OkxMirrorAPIKey = os.Getenv(env_OKX_MIRROR_API_KEY)
}

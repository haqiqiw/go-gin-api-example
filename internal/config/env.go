package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	AppName         string
	AppPort         int
	AppReadTimeout  int
	AppWriteTimeout int

	DBHost            string
	DBPort            string
	DBName            string
	DBUsername        string
	DBPassword        string
	DBConnMaxLifetime int
	DBMaxOpenConn     int
	DBMaxIdleConn     int

	RedisHost  string
	RedistPort string
	RedistDB   int

	JWTSecretKey string

	KafkaBrokerHost          string
	KafkaConsumerGroup       string
	KafkaAutoOffsetReset     string
	KafkaTopicUserRegistered string
}

func NewEnv() (*Env, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	cfg := &Env{
		AppName:         getEnvString("APP_NAME", "api-example"),
		AppPort:         getEnvInt("APP_PORT", 8500),
		AppReadTimeout:  getEnvInt("APP_READ_TIMEOUT", 60),
		AppWriteTimeout: getEnvInt("APP_WRITE_TIMEOUT", 60),

		DBHost:            getEnvString("DATABASE_HOST", "127.0.0.1"),
		DBPort:            getEnvString("DATABASE_PORT", "3306"),
		DBName:            getEnvString("DATABASE_NAME", ""),
		DBUsername:        getEnvString("DATABASE_USERNAME", ""),
		DBPassword:        getEnvString("DATABASE_PASSWORD", ""),
		DBConnMaxLifetime: getEnvInt("DATABASE_CONN_MAX_LIFETIME", 180),
		DBMaxOpenConn:     getEnvInt("DATABASE_MAX_OPEN_CONN", 10),
		DBMaxIdleConn:     getEnvInt("DATABASE_MAX_IDLE_CONN", 10),

		RedisHost:  getEnvString("REDIST_HOST", "127.0.0.1"),
		RedistPort: getEnvString("REDIS_PORT", "6379"),
		RedistDB:   getEnvInt("REDIS_DB", 0),

		JWTSecretKey: getEnvString("JWT_SECRET_KEY", ""),

		KafkaBrokerHost:          getEnvString("KAFKA_BROKER_HOST", "127.0.0.1:9092"),
		KafkaConsumerGroup:       getEnvString("KAFKA_CONSUMER_GROUP", "api-example"),
		KafkaAutoOffsetReset:     getEnvString("KAFKA_AUTO_OFFSET_RESET", "latest"),
		KafkaTopicUserRegistered: getEnvString("KAFKA_TOPIC_USER_REGISTERED", "user-registered"),
	}

	return cfg, nil
}

func getEnvString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		pVal, err := strconv.Atoi(val)
		if err != nil {
			return defaultVal
		}
		return pVal
	}

	return defaultVal
}

package env

import (
	"context"
	"strconv"
	_ "github.com/alexa-infra/kalendar/internal/context"
)

func GetString(ctx context.Context, key, defaultValue string) string {
	lookupEnv := GetContextLookupEnv(ctx)
	value, exists := lookupEnv(key)
	if !exists {
		return defaultValue
	}

	return value
}

func GetInt(ctx context.Context, key string, defaultValue int) int {
	lookupEnv := GetContextLookupEnv(ctx)
	value, exists := lookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}

	return intValue
}

func GetBool(ctx context.Context, key string, defaultValue bool) bool {
	lookupEnv := GetContextLookupEnv(ctx)
	value, exists := lookupEnv(key)
	if !exists {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}

	return boolValue
}

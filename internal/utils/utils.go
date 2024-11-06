package utils

import (
	"log"
	"os"
	"strings"
)

// func EnsureEnvironment(key string, loggger *zerolog.Logger) string {
// 	envVal, isSet := os.LookupEnv(key)
// 	if !isSet || envVal == "" {
// 		loggger.Fatal().Msgf("%s environment variable not set or empty", key)
// 	}
// 	return envVal
// }

// func RetrieveHost(logger *zerolog.Logger) string {
// 	for _, envVar := range os.Environ() {
// 		key := strings.Split(envVar, "=")[0]
// 		val := strings.Split(envVar, "=")[1]
// 		if strings.Contains(key, "HOST_ENDPOINT_") {
// 			return val
// 		}
// 	}
// 	logger.Warn().Msg("HOST_ENDPOINT_XXXX variable not found!")
// 	return EnsureEnvironment("HOST", logger)
// }

func GetEnv(key string, logger *log.Logger) string {
	envVal, isSet := os.LookupEnv(key)
	if !isSet || envVal == "" {
		logger.Fatalf("%s environment variable not set or empty", key)
	}
	return envVal
}

func RetrieveHost(logger *log.Logger) string {
	for _, envVar := range os.Environ() {
		key := strings.Split(envVar, "=")[0]
		val := strings.Split(envVar, "=")[1]
		if strings.Contains(key, "HOST_ENDPOINT_") {
			return val
		}
	}
	logger.Println("HOST_ENDPOINT_XXXX variable not found, using HOST variable")
	return GetEnv("HOST", logger)
}

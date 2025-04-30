package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
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
	if key == "APPLICATION_TYPE" {
		return "custom"
	}
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
		if strings.Contains(key, "HOST_ENDPOINT") {
			return val
		}
	}
	logger.Println("HOST_ENDPOINT_XXXX variable not found, using HOST variable")
	return GetEnv("HOST", logger)
}

func DoPostRequest(url string, ipAddress string, logger *log.Logger) error {
	body := fmt.Sprintf("ip: %s", ipAddress)
	logger.Println("body:")
	logger.Printf("\t%s\n", body)
	logger.Println("url:", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func DebugEnv(logger *log.Logger) {
	logger.Println("Environment variables:")
	for _, envVar := range os.Environ() {
		logger.Println(envVar)
	}
}

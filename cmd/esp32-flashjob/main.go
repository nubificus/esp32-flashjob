package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/nubificus/esp32-flashjob/internal/utils"
	oci "github.com/nubificus/esp32-flashjob/pkg/firmware"
)

const (
	DefaultOS    string = "custom"
	OTAAgentPath string = "/usr/local/bin/ota-agent"
)

var logger = log.Default()

type OTAConfig struct {
	firmware    *oci.OCIFirmware
	host        string
	device      string
	application string
	version     string
}

func newOTAConfig() *OTAConfig {
	return &OTAConfig{}
}

func main() {
	logger.Println("esp32-ota initialized")
	utils.DebugEnv(logger)
	jobConfig := newOTAConfig()
	jobConfig.device = utils.GetEnv("DEVICE", logger)
	// quick fix for inconsistent device name
	if jobConfig.device == "esp32-s3" {
		jobConfig.device = "esp32s3"
	}
	jobConfig.host = utils.RetrieveHost(logger)
	jobConfig.application = utils.GetEnv("APPLICATION_TYPE", logger)
	jobConfig.version = utils.GetEnv("VERSION", logger)
	diceAuthServer := utils.GetEnv("DICE_AUTH_SERVICE_SERVICE_HOST", logger)
	jobConfig.firmware = oci.NewOCIFirmware(utils.GetEnv("FIRMWARE", logger))
	logger.Println("Parsed job options")
	logger.Printf("\t- Host: %s", jobConfig.host)
	logger.Printf("\t- Device: %s", jobConfig.device)
	logger.Printf("\t- Application: %s", jobConfig.application)
	logger.Printf("\t- Version: %s", jobConfig.version)
	logger.Printf("\t- Target Firmware: %s", jobConfig.firmware.Name())
	logger.Printf("\t- Target Version: %s", jobConfig.firmware.Version())
	logger.Printf("\t- Dice Auth Host: %s", diceAuthServer)

	agentIP := utils.GetEnv("EXTERNAL_IP", logger)
	serverCRT := "/ota/certs/server.crt"
	serverKey := "/ota/certs/server.key"

	err := jobConfig.firmware.DownloadWithPlatform(jobConfig.device, DefaultOS)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Printf("Firmware downloaded at %s", jobConfig.firmware.Destination())

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		time.Sleep(2 * time.Second)
		defer wg.Done()
		logger.Println("Requesting OTA initialization for agent", agentIP)
		err = utils.DoPostRequest(fmt.Sprintf("http://%s/update", jobConfig.host), agentIP, logger)
		if err != nil {
			logger.Fatalf("Error performing POST request: %v\n", err)
		}
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		cmd := exec.Command(OTAAgentPath)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("NEW_FIRMWARE_PATH=%s", jobConfig.firmware.Destination()),
			fmt.Sprintf("DICE_AUTH_URL=http://%s:8000", diceAuthServer),
			fmt.Sprintf("SERVER_CRT_PATH=%s", serverCRT),
			fmt.Sprintf("SERVER_KEY_PATH=%s", serverKey),
		)
		logger.Println("Executing ota-agent with env")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Fatalf("Error creating stdout pipe: %v\n", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			logger.Fatalf("Error creating stderr pipe: %v\n", err)
		}
		if err := cmd.Start(); err != nil {
			logger.Fatalf("Error starting ota-agent: %v\n", err)
		}

		// Use a WaitGroup to wait for both goroutines to finish
		var logWg sync.WaitGroup
		logWg.Add(2)

		// go io.Copy(os.Stdout, stdout)
		go func() {
			defer logWg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				logger.Printf("[AGENT STDOUT] %s", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				logger.Printf("Error reading stdout: %v", err)
			}
		}()
		// go io.Copy(os.Stderr, stderr)
		go func() {
			defer logWg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				logger.Printf("[AGENT STDERR] %s", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				logger.Printf("Error reading stderr: %v", err)
			}
		}()
		if err := cmd.Wait(); err != nil {
			logger.Fatalf("Error waiting for ota-agent: %v\n", err)
		}
		logWg.Wait()
		logger.Println("OTA Agent exited gracefully")
	}()

	wg.Wait()
	logger.Println("Done!")
}

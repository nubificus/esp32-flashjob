package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/nubificus/esp32-sota/internal/utils"
	oci "github.com/nubificus/esp32-sota/pkg/firmware"
)

const (
	DefaultOS      string = "custom"
	OTAAgentPath   string = "/ota-agent"
	DeviceInfoFile string = "/dev_info.txt"
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

	jobConfig := newOTAConfig()
	jobConfig.device = utils.GetEnv("DEVICE", logger)
	jobConfig.host = utils.RetrieveHost(logger)
	jobConfig.application = utils.GetEnv("APPLICATION_TYPE", logger)
	jobConfig.version = utils.GetEnv("VERSION", logger)
	jobConfig.firmware = oci.NewOCIFirmware(utils.GetEnv("FIRMWARE", logger))
	logger.Println("Parsed job options")
	logger.Printf("\t- Host: %s", jobConfig.host)
	logger.Printf("\t- Device: %s", jobConfig.device)
	logger.Printf("\t- Application: %s", jobConfig.application)
	logger.Printf("\t- Version: %s", jobConfig.version)
	logger.Printf("\t- Target Firmware: %s", jobConfig.firmware.Name())
	logger.Printf("\t- Target Version: %s", jobConfig.firmware.Version())

	ownIP := utils.GetEnv("AGENT_IP", logger)

	err := jobConfig.firmware.DownloadWithPlatform(jobConfig.device, DefaultOS)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Printf("Firmware downloaded at %s", jobConfig.firmware.Destination())

	// Create the file
	file, err := os.Create(DeviceInfoFile)
	if err != nil {
		logger.Fatalf("Error creating file: %v\n", err)
	}

	// TODO: We need to populate the file containing the device info: MAC, App Hash and Bootloader Hash

	// Close the file immediately
	err = file.Close()
	if err != nil {
		logger.Fatalf("Error closing file: %v\n", err)
	}

	logger.Println("Requesting OTA initialization for agent ", ownIP)
	err = utils.DoPostRequest(fmt.Sprintf("http://%s/update", jobConfig.host), ownIP)
	if err != nil {
		logger.Fatalf("Error closing file: %v\n", err)
	}
	// TODO: we need to find a way to set the following inside the Pod
	// SERVER_CRT_PATH: the certificate that will be used by the server for the networking operations
	// SERVER_KEY_PATH: the correspondent private key

	cmd := exec.Command(OTAAgentPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("NEW_FIRMWARE_PATH=%s", jobConfig.firmware.Destination()),
		fmt.Sprintf("DEV_INFO_PATH=%s", DeviceInfoFile),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("Error: %v\n", err)
	}

	// Print output
	logger.Printf("OTA AgentOutput:\n%s\n", output)
}

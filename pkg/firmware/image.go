package firmware

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

// TODO: We make the assumption that the firmware blob will always be placed under "/firmware" directory
// and it will always be the only file under that directory.
// We should explore if there are any more suitable alternatives.
const FirmwareDir = "firmware"

type OCIFirmware struct {
	name         string // name of the firmware
	version      string // version of the firmware
	url          string // the URL of the OCI image containing the firmware
	firmwarePath string // the path of the firmware inside the OCI image's rootfs
	destination  string // the local path of the extracted firmware
}

func NewOCIFirmware(image string) *OCIFirmware {
	o := &OCIFirmware{}

	parts := strings.Split(image, ":")
	o.version = parts[len(parts)-1]

	parts = strings.Split(image, "/")
	o.name = parts[len(parts)-1]
	o.name = strings.ReplaceAll(o.name, ":", "")
	o.name = strings.ReplaceAll(o.name, o.version, "")
	o.url = image
	o.destination = ""
	o.firmwarePath = ""
	return o
}

// Name returns the name of the firmware OCI image
func (o OCIFirmware) Name() string {
	return o.name
}

// Version returns the version of the firmware.
// Currently, this is extracted from the OCI image tag provided.
func (o OCIFirmware) Version() string {
	return o.version
}

// URL returns the URL of the OCI image containing the firmware
func (o OCIFirmware) URL() string {
	return o.url
}

// FirmwarePath returns the path of the firmware inside the OCI image's rootfs.
// If the image is not yet unpacked, it returns an empty string
func (o OCIFirmware) FirmwarePath() string {
	return o.firmwarePath
}

// Destination returns the local path of the extracted firmware.
func (o OCIFirmware) Destination() string {
	return o.destination
}

// Download attempts to download and unpack the OCI image provided by the url.
func (o *OCIFirmware) Download() error {
	// Pull image from registry
	img, err := crane.Pull(o.url)
	if err != nil {
		return err
	}
	// Load flattened image filesystem
	imageFS := mutate.Extract(img)
	defer imageFS.Close()
	tarReader := tar.NewReader(imageFS)

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "firmware*")
	if err != nil {
		return err
	}
	fmt.Println(tmpDir)
	// Iterate over the TAR archive
	for {
		// Read the next entry from the tar archive
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of the tar archive
			break
		}
		if err != nil {
			return err
		}
		// TODO: This is a very relaxed way of determining we found the firmware file.
		// We should implement a stricter check.
		if strings.Contains(header.Name, FirmwareDir) && header.Name != FirmwareDir {
			fullPath := filepath.Join("/", header.Name)
			o.firmwarePath = fullPath
			firmwareName := filepath.Base(fullPath)
			newFilePath := filepath.Join(tmpDir, filepath.Base(firmwareName))
			o.destination = newFilePath
			outFile, err := os.Create(newFilePath)
			if err != nil {
				return err
			}
			fmt.Println(outFile)
			defer outFile.Close()
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
			return nil
		}
	}

	// Remove the temp directory if the firmware was not found.
	os.RemoveAll(tmpDir)
	return fmt.Errorf("firmware not found in %s rootfs", o.URL())
}

// Download attempts to download and unpack the OCI image provided by the url.
func (o *OCIFirmware) DownloadWithPlatform(architecture, operatingSystem string) error {
	option := crane.WithPlatform(&v1.Platform{
		Architecture: architecture,
		OS:           operatingSystem,
	})
	// Pull image from registry
	img, err := crane.Pull(o.url, option)
	if err != nil {
		return err
	}
	// Load flattened image filesystem
	imageFS := mutate.Extract(img)
	defer imageFS.Close()
	tarReader := tar.NewReader(imageFS)

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "firmware*")
	if err != nil {
		return err
	}
	// Iterate over the TAR archive
	for {
		// Read the next entry from the tar archive
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of the tar archive
			break
		}
		if err != nil {
			return err
		}
		// TODO: This is a very relaxed way of determining we found the firmware file.
		// We should implement a stricter check.
		if strings.Contains(header.Name, FirmwareDir) && header.Name != FirmwareDir {
			fullPath := filepath.Join("/", header.Name)
			o.firmwarePath = fullPath
			firmwareName := filepath.Base(fullPath)
			newFilePath := filepath.Join(tmpDir, filepath.Base(firmwareName))
			o.destination = newFilePath
			outFile, err := os.Create(newFilePath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
			return nil
		}
	}

	// Remove the temp directory if the firmware was not found.
	os.RemoveAll(tmpDir)
	return fmt.Errorf("firmware not found in %s rootfs", o.URL())
}

// Clear deletes any artifacts created during downloading and extraction
func (o *OCIFirmware) Clear() error {
	if o.destination == "" {
		return fmt.Errorf("destination was not set, cannot clear")
	}
	return os.RemoveAll(filepath.Dir(o.destination))
}

# esp32-sota

Secure OTA Firmware Upgrades for ESP32

```bash
git clone git@github.com:nubificus/esp32-sota.git
git submodule update --init --recursive
docker build --push -t harbor.nbfc.io/nubificus/iot/esp32-sota:1252 .
```

To run it "locally":

```bash
tee ota.env > /dev/null << 'EOT'
DEVICE=esp32
HOST=192.168.11.40
FIRMWARE=docker.io/gntouts/esp32-thermo-secure-firmware:0.0.2
APPLICATION_TYPE=null
VERSION=null
AGENT_IP=192.168.5.9
DEV_INFO_PATH=/ota/boards.txt
SERVER_CRT_PATH=/ota/certs/server.crt
SERVER_KEY_PATH=/ota/certs/server.key
EOT

sudo nerdctl --address /run/k3s/containerd/containerd.sock run --network host --rm -ti --env-file ota.env harbor.nbfc.io/nubificus/iot/esp32-sota:0.1.1-static /esp32-sota
```

## build image from binary

```bash
TARGET_ARCH=esp32s3

docker buildx build --platform custom/$TARGET_ARCH -t harbor.nbfc.io/nubificus/iot/esp32-resnet:1-$TARGET_ARCH -f Dockerfile.res . --push --provenance false
docker manifest create harbor.nbfc.io/nubificus/iot/esp32-resnet:1 \
  --amend harbor.nbfc.io/nubificus/iot/esp32-resnet:1-$TARGET_ARCH
docker manifest push harbor.nbfc.io/nubificus/iot/esp32-resnet:1


docker buildx build --platform custom/esp32s2 -t harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0-esp32s2 --build-arg DEVICE=esp32s2 . --push --provenance false
docker buildx build --platform custom/esp32s3 -t harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0-esp32s3 --build-arg DEVICE=esp32s3 . --push --provenance false

docker manifest create harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0 \
  --amend harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0-esp32 \
  --amend harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0-esp32s2 \
  --amend harbor.nbfc.io/nubificus/iot/esp32-thermo-firmware:0.2.0-esp32s3

```

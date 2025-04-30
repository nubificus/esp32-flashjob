# esp32-flashjob

Secure OTA Firmware Upgrades for ESP32

```bash
git clone git@github.com:nubificus/esp32-flashjob.git
git submodule update --init --recursive
docker build --push -t harbor.nbfc.io/nubificus/iot/esp32-flashjob:1252 .
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

sudo nerdctl --address /run/k3s/containerd/containerd.sock run --network host --rm -ti --env-file ota.env harbor.nbfc.io/nubificus/iot/esp32-flashjob:0.1.1-static /esp32-flashjob
```

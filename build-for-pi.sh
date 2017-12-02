#!/bin/bash

PROJECT_LOCATION="github.com/skilia/photobooth"

PACKAGE_NAME="photobooth.tar.gz"

SSH_HOST="192.168.0.61"
SSH_USER="pi"
SSH_PATH="/home/pi/"
TMP_DIR="/tmp/${PACKAGE_NAME}"

# Sets build env for Pi 2
env GOOS=linux GOARCH=arm GOARM=6 revel package "${PROJECT_LOCATION}"

# Pushes file to Pi 2
rsync --progress "${PACKAGE_NAME}" "${SSH_USER}@${SSH_HOST}:${TMP_DIR}"

# Extracts packed combiled version
ssh "${SSH_USER}@${SSH_HOST}" -- mkdir -p "${SSH_PATH}" && tar -C "${SSH_PATH}" -xzf "${TMP_DIR}"

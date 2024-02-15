#!/bin/bash

MAINTAINER="Peltoche <peltoche@halium.fr>"
PROJECT_NAME="duckcloud"
PROJECT_DESC="Small but full-featured DAV* server"

OUTPUT_DIR=${OUTPUT_DIR:="builds"}

DPKG_VERSION=${DPKG_VERSION:="unknown"}
DPKG_DIR=${DPKG_DIR:="${PWD}/dpkg"}
DPKG_ARCH=${DPKG_ARCH:="unknown"}
DPKG_NAME="${DPKG_NAME:="${PROJECT_NAME}_${DPKG_VERSION}_${DPKG_ARCH}.deb"}"

# README and LICENSE
install -Dm644 ./LICENSE "${DPKG_DIR}/usr/share/licences/${PROJECT_NAME}/LICENSE"

# Binary
install -Dm755 './duckcloud' "${DPKG_DIR}/usr/bin/${PROJECT_NAME}"

# Systemd unit file
install -Dm755 './resources/duckcloud.service' "${DPKG_DIR}/usr/lib/systemd/system/${PROJECT_NAME}.service"

# Config file
install -Dm644 './resources/var_file' "${DPKG_DIR}/etc/${PROJECT_NAME}/var_file"

# Default data dir
install -Dm755 -d "${DPKG_DIR}/usr/share/${PROJECT_NAME}"

# Post install script
install -Dm755 './scripts/generate_credentials.sh' "${DPKG_DIR}/DEBIAN/postinst"

# Control file
mkdir -p "${DPKG_DIR}/DEBIAN"
printf '%s\n' \
	"Package: ${PROJECT_NAME}" \
	"Version: ${DPKG_VERSION}" \
	"Section: utils" \
	"Priority: optional" \
	"Maintainer: ${MAINTAINER}" \
	"Architecture: ${DPKG_ARCH}" \
	"Provides: ${PROJECT_NAME}" \
	"Description: ${PROJECT_DESC}" \
	" Duckcloud aims to propose an easy solution to host a little cloud at home. It try to be as simple as possible for both the users and the administrator." >"${DPKG_DIR}/DEBIAN/control"

# Config file
printf '%s\n' \
	"/etc/${PROJECT_NAME}/var_file" >"${DPKG_DIR}/DEBIAN/conffiles"

# Copyright File

echo -e "\n\n#####\n## Control file\n#####\n"
cat "${DPKG_DIR}/DEBIAN/control"
echo -e "\n\n#####\n## End Control file\n#####\n"

# Build dpkg
mkdir -p ./builds/
fakeroot dpkg-deb --build "${DPKG_DIR}" "./builds/${DPKG_NAME}"

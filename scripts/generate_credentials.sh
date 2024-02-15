#!/bin/bash

PASSWORDFILE="/etc/duckcloud/password.cred"
TMPFILE="/tmp/password.txt"

if [ -f "$PASSWORDFILE" ]; then
	echo "$PASSWORDFILE already exists, do nothing"
else
	echo "Generate the password file at $PASSWORDFILE"
	openssl rand -hex 32 | head -n1 >"$TMPFILE"
	sudo systemd-creds encrypt --name=password "$TMPFILE" "$PASSWORDFILE"
	shred -u "$TMPFILE"
fi

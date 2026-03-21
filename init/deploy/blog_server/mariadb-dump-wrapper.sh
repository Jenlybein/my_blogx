#!/bin/sh
set -eu

if command -v mariadb-dump >/dev/null 2>&1; then
	exec "$(command -v mariadb-dump)" --skip-ssl-verify-server-cert "$@"
fi

exec "$(command -v mysqldump)" --skip-ssl-verify-server-cert "$@"

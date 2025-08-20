#!/bin/sh
BIN_PREFIX="${BIN_PREFIX:-sifeucr}"
DIST_FOLDER="${DIST_FOLDER:-dist}"
GOARCH_LIST="${GOARCH_LIST:-amd64 arm64}"
GOOS_LIST="${GOOS_LIST:-linux windows}"
COPY_FILES="${COPY_FILES:-extra/sifeucr.service extra/config.env}"

if [ "${DIST_FOLDER%/*}" != "." ]; then
	mkdir -p "${DIST_FOLDER%/*}" || exit 1
fi

for f in $COPY_FILES; do
	if ! [ -f "${DIST_FOLDER%/}/$f" ]; then
		cp "$f" "${DIST_FOLDER%/}/" || exit 1
	fi
done

for GOARCH in $GOARCH_LIST; do
	for GOOS in $GOOS_LIST; do
		if [ "$GOOS" = "windows" ]; then suffix=".exe" ; else suffix="" ; fi
		CGO_ENABLED=0 \
			GOARCH="$GOARCH" \
			GOOS="$GOOS" \
			go build \
			-ldflags "-s -w" \
			-o "${DIST_FOLDER%/}/$BIN_PREFIX-$GOARCH-$GOOS$suffix" || exit 1
	done
done

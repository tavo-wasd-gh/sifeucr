#!/bin/sh
PREFIX="/usr/local/bin"
SYSTEMD_PREFIX="/etc/systemd/system"

SIFEUCR_GITHUB="https://github.com/tavo-wasd-gh/sifeucr"
SIFEUCR_HOME="/var/lib/sifeucr"
SIFEUCR_CONFIG_HOME="/etc/sifeucr"
SIFEUCR_BIN="sifeucr-amd64-linux"

_latest_remote() {
	# shellcheck disable=SC1083
	release_url="$(curl -Ls -o /dev/null -w %{url_effective} "$SIFEUCR_GITHUB"/releases/latest)"
	if [ "$?" -eq 1 ]; then
		return 1
	fi
	echo "${release_url##*/v}"
}

if [ "$(whoami)" != "root" ]; then
	echo "error: must be root user"
	exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
	echo "error: curl not installed"
	exit 1
fi

if systemctl is-active --quiet sifeucr; then
	echo "error: sifeucr is running, stop before reinstalling"
	exit 1
fi

if ! id sifeucr >/dev/null 2>&1; then
	echo "adding user 'sifeucr' with home directory '$SIFEUCR_HOME'"

	if ! useradd -m -d "$SIFEUCR_HOME" sifeucr; then
		echo "error adding user 'sifeucr' with home directory '$SIFEUCR_HOME'"
		exit 1
	fi
fi

for d in "$SIFEUCR_HOME" "$SIFEUCR_CONFIG_HOME"; do
	if ! [ -d "$d" ]; then
		echo "creating directory '$d'"

		if ! mkdir -p "$d"; then
			echo "error creating directory '$d'"
		fi
	fi
done

version="$(_latest_remote)"
if [ "$?" != 0 ]; then
	echo "error getting latest version"
	exit 1
fi
download_url="$SIFEUCR_GITHUB/releases/download/v$version"

echo "downloading $download_url/$SIFEUCR_BIN into $PREFIX/sifeucr ..."
if ! curl -sLo "$PREFIX"/sifeucr "$download_url"/"$SIFEUCR_BIN"; then
	echo "error downloading"
fi

echo "downloading $download_url/sifeucr.service into $SYSTEMD_PREFIX/sifeucr.service ..."
if ! curl -sLo "$SYSTEMD_PREFIX"/sifeucr.service "$download_url"/sifeucr.service; then
	echo "error downloading"
fi
systemctl daemon-reload 2>/dev/null

if ! [ -f "$SIFEUCR_CONFIG_HOME"/config.env ]; then
	echo "downloading $download_url/config.env into $SIFEUCR_CONFIG_HOME/config.env ..."

	if ! curl -sLo "$SIFEUCR_CONFIG_HOME"/config.env "$download_url"/config.env; then
		echo "error downloading"
	fi
fi

# Custom configuration (or edit /etc/sifeucr/config.env after startup):
#cp /path/to/config.env /etc/sifeucr/config.env
# Custom database
#cp /path/to/init.db /var/lib/sifeucr/db.db
# Load previous files
#cp -r /path/to/datafolder /var/lib/sifeucr/data
# Custom binary
#cp /path/to/sifeucrbinary /usr/local/bin/sifeucr

echo "setting up permissions ..."
for d in "$SIFEUCR_HOME" "$SIFEUCR_CONFIG_HOME"; do
	chown -R sifeucr:sifeucr "$d"
	find "$d" -type d -exec chmod 700 {} \;
	find "$d" -type f -exec chmod 600 {} \;
done
unset d

echo "done!"

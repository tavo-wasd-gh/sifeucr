#!/bin/sh

MARKED_FOR_DELETE="/etc/sifeucr /var/lib/sifeucr extra/sifeucr-install.sh"

for d in $MARKED_FOR_DELETE; do
	rm -rf "$d"
done

userdel sifeucr

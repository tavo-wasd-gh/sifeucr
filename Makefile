release: public views templates default.db .env.example
	rm -rf release sifeucr.tar.gz
	mkdir -p release/views release/templates
	cp -r views/*.html release/views/ && cp -r templates/*.html release/templates && cp -r public release/ || rm -rf release
	cp default.db release/ || rm -rf release
	cp docs/sifeucr.service release/ || rm -rf release
	cp .env.example release/ || rm -rf release
	cp Makefile release/ || rm -rf release
	podman build --rm -v $(shell pwd)/release:/release \
		-t sifeucr-release -f Dockerfile.release . || rm -rf release
	cp -r release sifeucr && \
	tar czf sifeucr.tar.gz sifeucr || rm -rf sifeucr.tar.gz release sifeucr
	rm -rf sifeucr

public:
	hugo

default.db: docs/schema.sql
	sqlite3 default.db ".read docs/schema.sql"

db.db: .data docs/schema.sql default.db
	cp default.db db.db
	sqlite3 db.db ".import .data/usuarios.csv usuarios --csv"
	sqlite3 db.db ".import .data/cuentas.csv cuentas --csv"
	sqlite3 db.db ".import .data/presupuestos.csv presupuestos --csv"

install:
	mkdir -p /usr/local/bin
	cp sifeucr /usr/local/bin/
	cp sifeucr.service /etc/systemd/system/
	@if ! id -u sifeucr >/dev/null 2>&1; then \
		echo "Creating user sifeucr"; \
		sudo useradd -m sifeucr; \
		fi
	mkdir -p /home/sifeucr/sifeucr
	cp .env.example /home/sifeucr/sifeucr/
	cp default.db /home/sifeucr/sifeucr/
	rm -rf /home/sifeucr/sifeucr/views && \
		cp -r views /home/sifeucr/sifeucr/
	rm -rf /home/sifeucr/sifeucr/templates && \
		cp -r templates /home/sifeucr/sifeucr/
	rm -rf /home/sifeucr/sifeucr/public && \
		cp -r public /home/sifeucr/sifeucr/

clean:
	rm -rf default.db release sifeucr.tar.gz tmp public sifeucr

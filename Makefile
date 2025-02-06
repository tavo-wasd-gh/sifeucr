release: views templates default.db .env.example
	rm -rf release sifeucr.tar.gz
	mkdir -p release
	cp -r views release/ && cp -r templates release/ || rm -rf release
	cp default.db release/ || rm -rf release
	cp .env.example release/ || rm -rf release
	podman build --rm -v $(shell pwd)/release:/release \
		-t sifeucr-release -f Dockerfile.release . || rm -rf release
	tar czf sifeucr.tar.gz release || rm -rf sifeucr.tar.gz release

default.db: .data docs/schema.sql
	sqlite3 default.db ".read docs/schema.sql" && \
	sqlite3 default.db ".import .data/usuarios.csv usuarios --csv" && \
	sqlite3 default.db ".import .data/cuentas.csv cuentas --csv" && \
	sqlite3 default.db ".import .data/presupuestos.csv presupuestos --csv"

clean:
	rm -rf default.db release sifeucr.tar.gz tmp public

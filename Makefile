release: views templates default.db .env.example
	rm -rf release sifeucr.tar.gz
	mkdir -p release/views release/templates
	cp -r views/*.html release/views/ && cp -r templates/*.html release/templates || rm -rf release
	cp default.db release/ || rm -rf release
	cp .env.example release/ || rm -rf release
	podman build --rm -v $(shell pwd)/release:/release \
		-t sifeucr-release -f Dockerfile.release . || rm -rf release
	tar czf sifeucr.tar.gz release || rm -rf sifeucr.tar.gz release

default.db: .data docs/schema.sql
	sqlite3 default.db ".read docs/schema.sql"

clean:
	rm -rf default.db release sifeucr.tar.gz tmp public

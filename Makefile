BIN = sifeucr
SRCDIR = server
SRC = ${SRCDIR}/main.go \
      # ${SRCDIR}/db.go \
      # ${SRCDIR}/auth.go \
      # ${SRCDIR}/bucket.go \

GO = go
GOFILES = ${SRCDIR}/go.sum ${SRCDIR}/go.mod
GOMODS = github.com/tavo-wasd-gh/gopdf \
         github.com/tavo-wasd-gh/gocors \

all: fmt ${BIN}

${BIN}: ${SRC} ${GOFILES}
	(cd ${SRCDIR} && ${GO} build -o ../${BIN})

fmt: ${SRC}
	@diff=$$(gofmt -d $^); \
	if [ -n "$$diff" ]; then \
		printf '%s\n' "$$diff"; \
		exit 1; \
	fi

${GOFILES}:
	(cd ${SRCDIR} && ${GO} mod init ${BIN})
	(cd ${SRCDIR} && ${GO} get ${GOMODS})

start: ${BIN}
	@./$< &

stop:
	-@pkill -SIGTERM ${BIN} || true

restart: stop start

clean-all: clean clean-mods clean-hugo

clean:
	rm -f ${BIN}

clean-mods:
	go clean -modcache
	rm -f ${SRCDIR}/go.*

clean-hugo:
	rm -f .hugo_build.lock
	rm -rf resources/
	rm -rf public/

.PHONY: test vet lint build examples clean docker-build docker-run

TAGS_NO_CGO := -tags='!focas_cgo'
TAGS_CGO    := -tags=focas_cgo

test:
	go test $(TAGS_NO_CGO) ./...

test-integration:
	go test $(TAGS_CGO) -tags=integration ./... -v

vet:
	go vet $(TAGS_NO_CGO) ./...

lint:
	staticcheck $(TAGS_NO_CGO) ./...

build:
	go build $(TAGS_NO_CGO) ./...

examples:
	@for d in examples/*/; do \
		echo "Building $$d ..."; \
		go build $(TAGS_NO_CGO) ./$$d; \
	done

clean:
	rm -rf bin/

# Fetch the x86-64 FOCAS shared library from the adapter-fanuc project.
# Required before docker-build if libfwlib32.so is not already present.
libfwlib32.so:
	cp ../../../work/CNC/adapter-fanuc/libfwlib32.so ./libfwlib32.so

docker-build: libfwlib32.so
	docker build --platform linux/amd64 -t go-focas .

# Usage: make docker-run EXAMPLE=02-read-position FOCAS_ADDR=192.168.1.100
docker-run:
	FOCAS_ADDR=$(FOCAS_ADDR) docker compose --profile $(or $(EXAMPLE),connect) up --build

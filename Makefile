

build:
	go build -o bin/stonkcritter ./cmd/stonkcritter

pkg: build
	mkdir -p dpkg/usr/bin
	cp bin/stonkcritter dpkg/usr/bin
	IAN_DIR=dpkg ian pkg
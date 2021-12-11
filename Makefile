

build:
	go build -o bin/stonkcritter ./cmd/politstonk

pkg: build
	mkdir -p dpkg/usr/bin
	cp bin/stonkcritter dpkg/usr/bin
	IAN_DIR=dpkg ian pkg
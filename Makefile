# This Makefile is needed to compile the Rust glue layer between Devzat and the
# Rust library Rustrict. To compile Devzat, simply run `make`. To install it as
# a Go binary, run `make install`.

all: devzat

GO_SRC := $(shell ls *.go)

devzat: $(GO_SRC) librustrict_devzat.a
	go build

librustrict_devzat.a: ./rustrict_devzat/src/lib.rs ./rustrict_devzat/Cargo.toml
	cd ./rustrict_devzat; \
	cargo build --lib --release; \
	cp target/release/librustrict_devzat.a ../; \
	cd ..

.PHONY: install
install: devzat
	go install

.PHONY: clean
clean:
	rm -rf devzat
	rm -rf librustrict_devzat.a
	cd ./rustrict_devzat; \
	cargo clean; \
	cd ..


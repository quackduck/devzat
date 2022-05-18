
GO_SRC := colors.go commands.go config.go games.go main.go rpc.go slack.go twitter.go util.go

devzat: $(GO_SRC) librustrict_devzat.a
	go build

librustrict_devzat.a: ./rustrict_devzat/src/lib.rs ./rustrict_devzat/Cargo.toml
	cd ./rustrict_devzat; \
	cargo build --lib --release; \
	cp target/release/librustrict_devzat.a ../; \
	cd ..

clean:
	rm -rf devzat
	rm -rf librustrict_devzat.a


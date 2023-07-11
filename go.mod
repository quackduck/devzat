module devzat

go 1.18

require (
	devzat/plugin v0.0.0-00010101000000-000000000000
	github.com/TwiN/go-away v1.6.10
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alecthomas/chroma v0.10.0
	github.com/bwmarrin/discordgo v0.27.2-0.20230704233747-e39e715086d2
	github.com/dghubble/go-twitter v0.0.0-20221104224141-912508c3888b
	github.com/dghubble/oauth1 v0.7.2
	github.com/gliderlabs/ssh v0.3.5
	github.com/jwalton/gchalk v1.3.0
	github.com/leaanthony/go-ansi-parser v1.6.1
	github.com/quackduck/go-term-markdown v0.14.2
	github.com/quackduck/term v0.0.0-20230512153006-5935fcd4d5e9
	github.com/shurcooL/tictactoe v0.0.0-20210613024444-e573ff1376a3
	github.com/slack-go/slack v0.12.2
	golang.org/x/image v0.9.0
	google.golang.org/grpc v1.56.2
	gopkg.in/yaml.v2 v2.4.0
)

require google.golang.org/genproto/googleapis/rpc v0.0.0-20230706204954-ccb25ca9f130 // indirect

replace devzat/plugin => ./plugin

require (
	github.com/MichaelMure/go-term-text v0.3.1 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/caarlos0/sshmarshal v0.1.0
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/dghubble/sling v1.4.1 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/eliukblau/pixterm v1.3.1 // indirect
	github.com/fatih/color v1.15.0
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gomarkdown/markdown v0.0.0-20230322041520-c84983bdbf2a // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jwalton/go-supportscolor v1.2.0 // indirect
	github.com/kyokomi/emoji/v2 v2.2.12 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	golang.org/x/crypto v0.11.0
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/term v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

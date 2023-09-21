package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigFile = "devzat.yml"
	DefaultDataDir    = "devzat-data"
	DefaultKeyFile    = "devzat-sshkey"
)

type ConfigType struct {
	Port        int               `yaml:"port"`
	AltPort     int               `yaml:"alt_port"`
	ProfilePort int               `yaml:"profile_port"`
	Scrollback  int               `yaml:"scrollback"`
	DataDir     string            `yaml:"data_dir"`
	KeyFile     string            `yaml:"key_file"`
	Admins      map[string]string `yaml:"admins"`
	Censor      bool              `yaml:"censor,omitempty"`
	Private     bool              `yaml:"private,omitempty"`
	Allowlist   map[string]string `yaml:"allowlist,omitempty"`

	IntegrationConfig string `yaml:"integration_config"`
}

type IntegrationsType struct {
	Twitter *TwitterInfo `yaml:"twitter"`
	Slack   *SlackInfo   `yaml:"slack"`
	Discord *DiscordInfo `yaml:"discord"`

	RPC *RPCInfo `yaml:"rpc"`
}

type TwitterInfo struct {
	ConsumerKey       string `yaml:"consumer_key"`
	ConsumerSecret    string `yaml:"consumer_secret"`
	AccessToken       string `yaml:"access_token"`
	AccessTokenSecret string `yaml:"access_token_secret"`
}

type SlackInfo struct {
	Token     string `yaml:"token"`
	ChannelID string `yaml:"channel_id"`
	Prefix    string `yaml:"prefix"`
}

type DiscordInfo struct {
	Token       string `yaml:"token"`
	ChannelID   string `yaml:"channel_id"`
	Prefix      string `yaml:"prefix"`
	CompactMode bool   `yaml:"compact_mode"`
}

type RPCInfo struct {
	Port int    `yaml:"port"`
	Key  string `yaml:"key"`
}

var (
	Config = ConfigType{
		Port:        2221,
		AltPort:     8080,
		ProfilePort: 5555,
		Scrollback:  16,
		DataDir:     DefaultDataDir,
		KeyFile:     DefaultKeyFile,

		IntegrationConfig: "",
	}

	Integrations = IntegrationsType{} // all nil

	Log *log.Logger
)

func init() {
	readConfig()
	setupLog()
	initBacklog()
	initIntegrations()
}

func readConfig() {
	cfgFile := os.Getenv("DEVZAT_CONFIG")
	if cfgFile == "" {
		cfgFile = DefaultConfigFile
	}

	var d []byte
	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found, so using the default one and writing it to " + cfgFile)

			d, err = yaml.Marshal(Config)
			errCheck(err)
			err = os.WriteFile(cfgFile, d, 0644)
		}
		errCheck(err)
	} else {
		d, err = os.ReadFile(cfgFile)
		errCheck(err)
		err = yaml.UnmarshalStrict(d, &Config)
		errCheck(err)
		fmt.Println("Config loaded from " + cfgFile)
	}

	err := os.MkdirAll(Config.DataDir, 0755)
	errCheck(err)
}

func setupLog() {
	logfile, err := os.OpenFile(Config.DataDir+string(os.PathSeparator)+"log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	errCheck(err)
	Log = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
}

func initBacklog() {
	Backlog = make([]backlogMessage, Config.Scrollback)
}

func initIntegrations() {
	if Config.IntegrationConfig != "" {
		d, err := os.ReadFile(Config.IntegrationConfig)
		errCheck(err)
		err = yaml.UnmarshalStrict(d, &Integrations)
		errCheck(err)

		if Integrations.Slack != nil {
			if Integrations.Slack.Prefix == "" {
				Integrations.Slack.Prefix = "Slack"
			}
			if sl := Integrations.Slack; sl.Token == "" || sl.ChannelID == "" {
				fmt.Println("error: Slack token or channel ID is missing")
				os.Exit(0)
			}
		}
		if Integrations.Discord != nil {
			if Integrations.Discord.Prefix == "" {
				Integrations.Discord.Prefix = "Discord"
			}
			if sl := Integrations.Discord; sl.Token == "" || sl.ChannelID == "" {
				fmt.Println("error: Discord token or channel ID is missing")
				os.Exit(0)
			}
		}
		if Integrations.Twitter != nil {
			if tw := Integrations.Twitter; tw.AccessToken == "" ||
				tw.AccessTokenSecret == "" ||
				tw.ConsumerKey == "" ||
				tw.ConsumerSecret == "" {
				fmt.Println("error: Twitter credentials are incomplete")
				os.Exit(0)
			}
		}

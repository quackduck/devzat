package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v2"
)

type ConfigType struct {
	Port        int               `yaml:"port" env:"PORT"`
	AltPort     int               `yaml:"alt_port" env:"ALT_PORT"`
	ProfilePort int               `yaml:"profile_port" env:"PROFILE_PORT"`
	Scrollback  int               `yaml:"scrollback" env:"SCROLLBACK"`
	DataDir     string            `yaml:"data_dir" env:"DATA_DIR"`
	KeyFile     string            `yaml:"key_file" env:"KEY_FILE"`
	Admins      map[string]string `yaml:"admins" env:"ADMINS"`
	Censor      bool              `yaml:"censor,omitempty" env:"CENSOR"`
	Private     bool              `yaml:"private,omitempty" env:"PRIVATE"`
	Allowlist   map[string]string `yaml:"allowlist,omitempty" env:"ALLOWLIST"`

	IntegrationConfig string           `yaml:"integration_config"`
	Integrations      IntegrationsType `yaml:"-" env:", prefix=INTEGRATION_"`
}

// IntegrationsType stores information needed by integrations.
// Code that uses this should check if fields are nil.
type IntegrationsType struct {
	// Twitter stores the information needed for the Twitter integration.
	// Check if it is enabled by checking if Twitter is nil.
	Twitter *TwitterInfo `yaml:"twitter" env:", prefix=TWITTER_, noinit"`
	// Slack stores the information needed for the Slack integration.
	// Check if it is enabled by checking if Slack is nil.
	Slack *SlackInfo `yaml:"slack" env:", prefix=SLACK_, noinit"`
	// Discord stores the information needed for the Discord integration.
	// Check if it is enabled by checking if Discord is nil.
	Discord *DiscordInfo `yaml:"discord" env:", prefix=DISCORD_, noinit"`

	RPC *RPCInfo `yaml:"rpc" env:", prefix=RPC_, noinit"`
}

type TwitterInfo struct {
	ConsumerKey       string `yaml:"consumer_key" env:"CONSUMER_KEY"`
	ConsumerSecret    string `yaml:"consumer_secret" env:"CONSUMER_SECRET"`
	AccessToken       string `yaml:"access_token" env:"ACCESS_TOKEN"`
	AccessTokenSecret string `yaml:"access_token_secret" env:"ACCESS_TOKEN_SECRET"`
}

type SlackInfo struct {
	// Token is the Slack API token
	Token string `yaml:"token" env:"TOKEN"`
	// ChannelID is the Slack channel to post to
	ChannelID string `yaml:"channel_id" env:"CHANNEL_ID"`
	// Prefix is the prefix to prepend to messages from Slack when rendered for SSH users
	Prefix string `yaml:"prefix" env:"PREFIX"`
}

type DiscordInfo struct {
	// Token is the Discord API token
	Token string `yaml:"token" env:"TOKEN"`
	// ChannelID is the ID of the channel to post to
	ChannelID string `yaml:"channel_id" env:"CHANNEL_ID"`
	// Prefix is the prefix to prepend to messages from Discord when rendered for SSH users
	Prefix string `yaml:"prefix" env:"PREFIX"`
	// Compact mode disables avatars to save vertical space
	CompactMode bool `yaml:"compact_mode" env:"COMPACT_MODE"`
}

type RPCInfo struct {
	Port int    `yaml:"port" env:"PORT"`
	Key  string `yaml:"key" env:"KEY"`
}

var (
	Config = ConfigType{ // first stores default config
		Port:        2221,
		AltPort:     8080,
		ProfilePort: 5555,
		Scrollback:  16,
		DataDir:     "devzat-data",
		KeyFile:     "devzat-sshkey",

		IntegrationConfig: "",
	}

	//Integrations = IntegrationsType{} // all nil

	Log *log.Logger
)

func init() {
	cfgFile := os.Getenv("DEVZAT_CONFIG")
	if cfgFile == "" {
		cfgFile = "devzat.yml"
	}

	Log = log.New(io.MultiWriter(os.Stdout, AdminLogWriter{}), "", log.Ldate|log.Ltime|log.Lshortfile)

	errCheck := func(err error) {
		if err != nil {
			Log.Println("err: " + err.Error())
			os.Exit(0) // match `return` behavior
		}
	}

	var d []byte
	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			Log.Println("Config file not found, so using the default one and writing it to " + cfgFile)

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
		Log.Println("Config loaded from " + cfgFile)
	}

	err := os.MkdirAll(Config.DataDir, 0755)
	errCheck(err)

	logfile, err := os.OpenFile(Config.DataDir+string(os.PathSeparator)+"log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	errCheck(err)
	Log.SetOutput(io.MultiWriter(logfile, os.Stdout, AdminLogWriter{}))

	if os.Getenv("PORT") != "" {
		Config.Port, err = strconv.Atoi(os.Getenv("PORT"))
		errCheck(err)
	}

	Backlog = make([]backlogMessage, Config.Scrollback)

	if Config.IntegrationConfig != "" {
		d, err = os.ReadFile(Config.IntegrationConfig)
		errCheck(err)
		err = yaml.UnmarshalStrict(d, &Config.Integrations)
		errCheck(err)

		if Config.Integrations.Slack != nil {
			if Config.Integrations.Slack.Prefix == "" {
				Config.Integrations.Slack.Prefix = "Slack"
			}
			if sl := Config.Integrations.Slack; sl.Token == "" || sl.ChannelID == "" {
				Log.Println("error: Slack token or channel ID is missing")
				os.Exit(0)
			}
		}
		if Config.Integrations.Discord != nil {
			if Config.Integrations.Discord.Prefix == "" {
				Config.Integrations.Discord.Prefix = "Discord"
			}
			if sl := Config.Integrations.Discord; sl.Token == "" || sl.ChannelID == "" {
				Log.Println("error: Discord token or channel ID is missing")
				os.Exit(0)
			}
		}
		if Config.Integrations.Twitter != nil {
			if tw := Config.Integrations.Twitter; tw.AccessToken == "" ||
				tw.AccessTokenSecret == "" ||
				tw.ConsumerKey == "" ||
				tw.ConsumerSecret == "" {
				Log.Println("error: Twitter credentials are incomplete")
				os.Exit(0)
			}
		}

		Log.Println("Integration config loaded from " + Config.IntegrationConfig)

		if os.Getenv("DEVZAT_OFFLINE_SLACK") != "" {
			Log.Println("Disabling Slack")
			Config.Integrations.Slack = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_DISCORD") != "" {
			Log.Println("Disabling Discord")
			Config.Integrations.Discord = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_TWITTER") != "" {
			Log.Println("Disabling Twitter")
			Config.Integrations.Twitter = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_RPC") != "" {
			Log.Println("Disabling RPC")
			Config.Integrations.RPC = nil
		}
		// Check for global offline for backwards compatibility
		if os.Getenv("DEVZAT_OFFLINE") != "" {
			Log.Println("Offline mode")
			Config.Integrations.Slack = nil
			Config.Integrations.Discord = nil
			Config.Integrations.Twitter = nil
			Config.Integrations.RPC = nil
		}
	}

	err = envconfig.Process(context.Background(), &Config)
	errCheck(err)

	slackInit()
	discordInit()
	twitterInit()
	rpcInit()
}

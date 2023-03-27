package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
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

// IntegrationsType stores information needed by integrations.
// Code that uses this should check if fields are nil.
type IntegrationsType struct {
	// Twitter stores the information needed for the Twitter integration.
	// Check if it is enabled by checking if Twitter is nil.
	Twitter *TwitterInfo `yaml:"twitter"`
	// Slack stores the information needed for the Slack integration.
	// Check if it is enabled by checking if Slack is nil.
	Slack *SlackInfo `yaml:"slack"`
	// Discord stores the information needed for the Discord integration.
	// Check if it is enabled by checking if Discord is nil.
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
	// Token is the Slack API token
	Token string `yaml:"token"`
	// ChannelID is the Slack channel to post to
	ChannelID string `yaml:"channel_id"`
	// Prefix is the prefix to prepend to messages from Slack when rendered for SSH users
	Prefix string `yaml:"prefix"`
}

type DiscordInfo struct {
	// Token is the Discord API token
	Token string `yaml:"token"`
	// ChannelID is the ID of the channel to post to
	ChannelID string `yaml:"channel_id"`
	// Prefix is the prefix to prepend to messages from Discord when rendered for SSH users
	Prefix string `yaml:"prefix"`
	// Compact mode disables avatars to save vertical space
	CompactMode bool `yaml:"compact_mode"`
}

type RPCInfo struct {
	Port int    `yaml:"port"`
	Key  string `yaml:"key"`
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

	Integrations = IntegrationsType{} // all nil

	Log *log.Logger
)

func init() {
	cfgFile := os.Getenv("DEVZAT_CONFIG")
	if cfgFile == "" {
		cfgFile = "devzat.yml"
	}

	errCheck := func(err error) {
		if err != nil {
			fmt.Println("err: " + err.Error())
			os.Exit(0) // match `return` behavior
		}
	}

	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found, so writing the default one to " + cfgFile)

			d, err := yaml.Marshal(Config)
			errCheck(err)
			err = os.WriteFile(cfgFile, d, 0644)
			errCheck(err)
			return
		}
		errCheck(err)
	}
	d, err := os.ReadFile(cfgFile)
	errCheck(err)
	err = yaml.UnmarshalStrict(d, &Config)
	errCheck(err)
	fmt.Println("Config loaded from " + cfgFile)

	err = os.MkdirAll(Config.DataDir, 0755)
	errCheck(err)

	logfile, err := os.OpenFile(Config.DataDir+string(os.PathSeparator)+"log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	errCheck(err)
	Log = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	if os.Getenv("PORT") != "" {
		Config.Port, err = strconv.Atoi(os.Getenv("PORT"))
		errCheck(err)
	}

	Backlog = make([]backlogMessage, Config.Scrollback)

	if Config.IntegrationConfig != "" {
		d, err = os.ReadFile(Config.IntegrationConfig)
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

		fmt.Println("Integration config loaded from " + Config.IntegrationConfig)

		if os.Getenv("DEVZAT_OFFLINE_SLACK") != "" {
			fmt.Println("Disabling Slack")
			Integrations.Slack = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_DISCORD") != "" {
			fmt.Println("Disabling Discord")
			Integrations.Discord = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_TWITTER") != "" {
			fmt.Println("Disabling Twitter")
			Integrations.Twitter = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_RPC") != "" {
			fmt.Println("Disabling RPC")
			Integrations.RPC = nil
		}
		// Check for global offline for backwards compatibility
		if os.Getenv("DEVZAT_OFFLINE") != "" {
			fmt.Println("Offline mode")
			Integrations.Slack = nil
			Integrations.Discord = nil
			Integrations.Twitter = nil
			Integrations.RPC = nil
		}
	}
	slackInit()
	discordInit()
	twitterInit()
	rpcInit()
}

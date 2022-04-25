package main

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type ConfigType struct {
	Port        int               `yaml:"port"`
	AltPort     int               `yaml:"alt_port"`
	ProfilePort int               `yaml:"profile_port"`
	DataDir     string            `yaml:"data_dir"`
	KeyFile     string            `yaml:"key_file"`
	Admins      map[string]string `yaml:"admins"`
	Censor      bool              `yaml:"censor,omitempty"`

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
	// Channel is the Slack channel to post to
	ChannelID string `yaml:"channel_id"`
	// Prefix is the prefix to prepend to messages from slack when rendered for SSH users
	Prefix string `yaml:"prefix"`
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
		DataDir:     "devzat-data",
		KeyFile:     "devzat-sshkey",
		Censor:      false,

		IntegrationConfig: "",
	}

	Integrations = IntegrationsType{
		Twitter: nil,
		Slack:   nil,
		RPC:     nil,
	}
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

	if os.Getenv("PORT") != "" {
		Config.Port, err = strconv.Atoi(os.Getenv("PORT"))
		errCheck(err)
	}

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
				fmt.Println("error: Slack token or Channel ID is missing")
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

		if Integrations.RPC != nil {
			if rpc := Integrations.RPC; rpc.Key == "" {
				fmt.Println("error: RPC key is missing")
				os.Exit(0)
			}
		}

		fmt.Println("Integration config loaded from " + Config.IntegrationConfig)

		if os.Getenv("DEVZAT_OFFLINE_SLACK") != "" {
			fmt.Println("Disabling Slack")
			Integrations.Slack = nil
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
			Integrations.Twitter = nil
			Integrations.RPC = nil
		}
	}
	slackInit()
	twitterInit()
	rpcInit()
}

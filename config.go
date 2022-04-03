package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type config struct {
	SSHPort     int `yaml:"ssh_port"`
	ProfilePort int `yaml:"profile_port"`

	DataDir   string `yaml:"data_dir"`
	KeyFile   string `yaml:"key_file"`
	CredsFile string `yaml:"creds_file"`
}

type secrets struct {
	Twitter twitterSecrets `yaml:"twitter"`
	Slack   slackSecrets   `yaml:"slack"`
}

type twitterSecrets struct {
	ConsumerKey       string `yaml:"consumer_key"`
	ConsumerSecret    string `yaml:"consumer_secret"`
	AccessToken       string `yaml:"access_token"`
	AccessTokenSecret string `yaml:"access_token_secret"`
}

type slackSecrets struct {
	// Token is the Slack API token
	Token string `yaml:"token"`
	// Channel is the Slack channel to post to
	ChannelID string `yaml:"channel_id"`
	// Prefix is the prefix to prepend to messages from slack when rendered for SSH users
	Prefix string `yaml:"prefix"`
}

var (
	// TODO: use this config!!

	Config = config{ // first stores default config
		2221,
		5555,
		"./devzat-data",
		"./devzat-sshkey",
		"./devzat-creds.json",
	}

	Secrets = secrets{
		twitterSecrets{
			"consumerkey",
			"consumersecret",
			"accesstoken",
			"accesstokensecret",
		},
		slackSecrets{
			"slacktoken",
			"channelid",
			"Slack",
		},
	}
)

func init() {
	cfgFile := os.Getenv("DEVZAT_CONFIG")
	if cfgFile == "" {
		cfgFile = "devzat-config.yml"
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
	d, err := ioutil.ReadFile(cfgFile)
	errCheck(err)
	err = yaml.Unmarshal(d, &Config)
	errCheck(err)
	fmt.Println("Config loaded from " + cfgFile)

	if Config.CredsFile != "" {
		d, err = ioutil.ReadFile(Config.CredsFile)
		errCheck(err)
		err = yaml.Unmarshal(d, &Secrets)
		errCheck(err)
		fmt.Println("Secrets loaded from " + Config.CredsFile)
	}
}

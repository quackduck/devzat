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

var (
	// TODO: use this config!!

	Config = config{ // first stores default config
		2221,
		5555,
		"./devzat-data",
		"./devzat-sshkey",
		"./devzat-creds.json",
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
			os.Exit(0)
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
	fmt.Println("Config loaded from "+cfgFile, Config)
}

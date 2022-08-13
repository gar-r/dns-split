package config

import (
	"encoding/json"
	"flag"
	"os"
)

const DefaultConfigLocation = "/etc/dns-split/config.json"

type Args struct {
	ConfigLocation string
	Verbose        bool
}

var DeclaredArgs = func() *Args {
	args := &Args{}
	flag.StringVar(&args.ConfigLocation, "config", DefaultConfigLocation,
		"specify the location of the config.json file")
	flag.BoolVar(&args.Verbose, "verbose", false,
		"enable verbose logging")
	return args
}

func Parse(args *Args) (*Config, error) {
	flag.Parse()
	dat, err := os.ReadFile(os.ExpandEnv(args.ConfigLocation))
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	return &cfg, err
}

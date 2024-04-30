package config

import (
	"encoding/json"
	"log"
	"os"
)

type Configs struct {
	TargetDir        string
	DatabaseLocation string
	NumThreads       int
	LogFilePath      string
	NewLogFileHours  int
}

const minNumThreads = 2
const maxNumThreads = 20
const defNumThreds = 5

const minLogFileHours = 1
const maxLogFileHours = 240
const defLogFileHours = 1

func LoadConfig() *Configs {
	flt, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("error opening config file")
		return nil
	}

	jsn := json.NewDecoder(flt)
	var cfg Configs
	err = jsn.Decode(&cfg)
	if err != nil {
		log.Fatalln("error decoding json config")
		return nil
	}
	if cfg.NumThreads < minNumThreads || cfg.NumThreads > maxNumThreads {
		cfg.NumThreads = defNumThreds
	}

	if cfg.NewLogFileHours < minLogFileHours || cfg.NewLogFileHours > maxLogFileHours {
		cfg.NewLogFileHours = defLogFileHours
	}

	return &cfg

}

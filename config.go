package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	General       ConfigGeneral
	DataProviders ConfigDataProviders
	Historians    ConfigHistorians
}

type ConfigDataProviders struct {
	CIPClass3 []*ConfigCIPClass3
}

type ConfigHistorians struct {
	Influx  []*ConfigHistorianInflux
	JSON    []*ConfigHistorianJSON
	Logging []*ConfigHistorianLogging
}

func (c *Config) Save(filename string) error {

	dat, err := json.MarshalIndent(*c, "", "	")
	if err != nil {
		return fmt.Errorf("problem marshaling: %w", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("problem saving config to %s: %w", filename, err)
	}
	defer f.Close()

	_, err = f.Write(dat)

	return err
}

func ConfigNew() Config {
	c := Config{}
	c.DataProviders.CIPClass3 = make([]*ConfigCIPClass3, 0)
	return c
}

func ConfigLoad(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("problem opening config file: %w", err)
	}
	defer f.Close()

	j := json.NewDecoder(f)
	c := Config{}
	err = j.Decode(&c)
	if err != nil {
		return Config{}, fmt.Errorf("problem parsing config file: %w", err)
	}

	return c, nil
}

type ConfigGeneral struct {
	Host         string
	Port         int
	RestartDelay time.Duration
}

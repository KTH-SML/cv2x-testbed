package main

import (
	"io/ioutil"
	"math/rand"

	"gopkg.in/yaml.v2"
)

func RandomBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		return []byte{}, err
	}
	return buf, nil
}

// Create or validate a checksum.
// To create checksum:
//
//	>>> c = checksum(b'...')
//
// To validate checksum against `c`:
//
//	>>> checksum(b'...', c) == 0
//
// Note: The checksum is valid if it returns 0.
func Checksum(data []byte, chk int) int {
	for _, b := range data {
		chk ^= int(b)
	}
	return chk
}

type Config struct {
	Host             string   `yaml:"host"`
	Publishers       []string `yaml:"publishers"`
	Subscribers      []string `yaml:"subscribers"`
	Services         []string `yaml:"services"`
	ServiceListeners []string `yaml:"service_listeners"`
}

func loadConf(file string) Config {
	config_bytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	var conf Config
	err = yaml.Unmarshal(config_bytes, &conf)
	if err != nil {
		panic(err)
	}
	return conf
}

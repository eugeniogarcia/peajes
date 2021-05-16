package main

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"isu"`

	Frecuencia int `yaml:"frecuencia"`

	ListaBatchs string `yaml:"lista"`
}

func cargar(conf *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(conf)
	if err != nil {
		processError(fmt.Errorf("no se ha especificado la lista de batchs: %s", err))
	}
	if conf.ListaBatchs == "" {
		processError(errors.New("no se ha especificado la lista de batchs"))
	}
	if conf.Server.Host == "" {
		processError(errors.New("no se ha especificado el host de ISU"))
	}
	if conf.Server.Port == "" {
		processError(errors.New("no se ha especificado el puerto de ISU"))
	}
	if conf.Frecuencia == 0 {
		conf.Frecuencia = 60
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

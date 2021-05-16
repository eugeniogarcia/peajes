package main

import (
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

    BatchIds []struct {
        batchid string `yaml:"batch"`
    } `yaml:"batchs"`

	ListaBatchs string `yaml:lista`
}

func cargar(conf *Config){
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(conf)
	if err != nil {
		processError(err)
	}
}

func processError(err error) {
    fmt.Println(err)
    os.Exit(2)
}

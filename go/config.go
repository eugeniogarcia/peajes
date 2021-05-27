package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eugeniogarcia/peajes/servicio"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"isu"`

	Puerto int `yaml:"escucha"`

	Frecuencia int `yaml:"frecuencia"`

	ListaBatchs string `yaml:"lista"`
	ListaCadena string `yaml:"cadenas"`

	Cadena int `yaml:"cadena"`

	ListaCadenas servicio.InformacionCadenas
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
	if conf.ListaCadena == "" {
		if conf.ListaBatchs == "" {
			processError(errors.New("no se ha especificado la lista de batchs"))
		}
	} else {
		conf.ListaCadenas = make(map[int][]string)

		resultado := ""
		if conf.Cadena == 0 {
			conf.Cadena = 120
		}
		if strings.ToLower(conf.ListaCadena) == "todas" {
			for i := 1; i <= conf.Cadena; i++ {
				conf.ListaCadenas[i] = make([]string, 5)
				for j := 0; j < 5; j++ {
					if i+j*conf.Cadena > 500 {
						continue
					} else {
						resultado += strconv.Itoa(i+j*conf.Cadena) + ","
						conf.ListaCadenas[i][j] = strconv.Itoa(i + j*conf.Cadena)
					}
				}
			}
		} else {
			valores := strings.Split(conf.ListaCadena, ",")
			for _, val := range valores {
				cad, _ := strconv.Atoi(val)
				conf.ListaCadenas[cad] = make([]string, 5)
				for j := 0; j < 5; j++ {
					if cad+j*conf.Cadena > 500 {
						continue
					} else {
						resultado += strconv.Itoa(cad+j*conf.Cadena) + ","
						conf.ListaCadenas[cad][j] = strconv.Itoa(cad + j*conf.Cadena)
					}
				}
			}
		}
		conf.ListaBatchs = resultado[:len(resultado)-1]
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
	if conf.Puerto == 0 {
		conf.Puerto = 9000
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

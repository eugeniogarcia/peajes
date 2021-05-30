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

	Frecuencia  int      `yaml:"frecuencia"`
	MaxJobs     int      `yaml:"max_jobs"`
	ListaBatchs []string `yaml:"lista"`
	ListaCadena []string `yaml:"cadenas"`

	Cadena int `yaml:"cadena"`

	Paciencia int `yaml:"paciencia"`

	ListaCadenas servicio.InformacionCadenas
}

const limiteJobs = 506

func cargar(conf *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	if conf.MaxJobs == 0 {
		conf.MaxJobs = limiteJobs
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(conf)
	if err != nil {
		processError(fmt.Errorf("no se ha especificado la lista de batchs: %s", err))
	}
	if conf.ListaCadena == nil {
		if conf.ListaBatchs == nil {
			processError(errors.New("no se ha especificado la lista de batchs"))
		}
	} else {
		conf.ListaCadenas = make(map[int][]string)

		if conf.Cadena == 0 {
			conf.Cadena = 120
		}
		if strings.ToLower(conf.ListaCadena[0]) == "todas" {
			for i := 1; i <= conf.Cadena; i++ {
				conf.ListaCadenas[i] = make([]string, 5)
				for j := 0; j < 5; j++ {
					if i+j*conf.Cadena > conf.MaxJobs {
						continue
					} else {
						conf.ListaBatchs = append(conf.ListaBatchs, strconv.Itoa(i+j*conf.Cadena))
						conf.ListaCadenas[i][j] = strconv.Itoa(i + j*conf.Cadena)
					}
				}
			}
		} else {
			for _, val := range conf.ListaCadena {
				cad, _ := strconv.Atoi(val)
				conf.ListaCadenas[cad] = make([]string, 5)
				for j := 0; j < 5; j++ {
					if cad+j*conf.Cadena > conf.MaxJobs {
						continue
					} else {
						conf.ListaBatchs = append(conf.ListaBatchs, strconv.Itoa(cad+j*conf.Cadena))
						conf.ListaCadenas[cad][j] = strconv.Itoa(cad + j*conf.Cadena)
					}
				}
			}
		}
	}
	if conf.ListaBatchs == nil {
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
	if conf.Paciencia == 0 {
		conf.Paciencia = 4
	}
	if conf.Puerto == 0 {
		conf.Puerto = 9000
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

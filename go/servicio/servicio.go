package servicio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var InformacionBatches Batches = Batches{Batches: make(map[string]*Batch), Frecuencia: 60}

type Runner struct {
	interrupt chan os.Signal
	complete  chan error
	wg        *sync.WaitGroup
}

func New(wg *sync.WaitGroup) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		wg:        wg,
	}
}

func (r *Runner) Start(entrada string, host string, puerto string, frecuencia int, totales *prometheus.GaugeVec, errores *prometheus.GaugeVec, activos *prometheus.GaugeVec) {

	//Nos subscribimos a las interrupciones del SSOO
	signal.Notify(r.interrupt, os.Interrupt)
	//Guardamos la frecuencia
	InformacionBatches.Frecuencia = frecuencia
	//Gauges de Prometheus
	InformacionBatches.Totales = totales
	InformacionBatches.Errores = errores
	InformacionBatches.Activos = activos

	//Prepara la entrada
	valores := strings.Split(entrada, ",")
	var input = make([]string, len(valores))
	for i, val := range valores {
		input[i] = fmt.Sprintf("\"batchid\":%s", val)
	}
	entrada_final := "{" + strings.Join(input, ",") + "}"

	//Arranca el servicio
	go func(entrada string) {
		//Notifica que se ha terminado el trabajo
		defer r.wg.Done()
		//Deja de antender a las interrupciones
		defer signal.Stop(r.interrupt)
		//Arranca el trabajo
		r.run(entrada, fmt.Sprintf("http://%s:%s/sap/bc/ZWS_JOB_MONITOR?sap-client=500", host, puerto), frecuencia)
	}(entrada_final)
}

func (r *Runner) run(entrada string, uri string, frecuencia int) {

	if monitorISU(uri, entrada) {
		log.Println("no queda nada por procesar en ISU")
		return
	}

	//Arranca el temporizador
	tick := time.NewTicker(time.Second * time.Duration(frecuencia))

	for {
		select {
		case <-r.interrupt:
			signal.Stop(r.interrupt)
			log.Println("se ha interrumpido la ejecución")
			return
		case <-tick.C:
			if monitorISU(uri, entrada) {
				log.Println("no queda nada por procesar en ISU")
				return
			}
		}
	}
}

func monitorISU(uri string, entrada string) bool {
	var jsonStr = []byte(entrada)
	req, err := http.NewRequest(http.MethodGet, uri, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
		return false
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println(err)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	defer resp.Body.Close()
	return procesa(body)

}

var paciencia int = 1

func procesa(body []byte) bool {
	var respuesta Respuesta
	json.Unmarshal(body, &respuesta)

	for _, medida := range respuesta {
		InformacionBatches.Add(medida.Batch, medida.Procesados, medida.Fallados, medida.Pendientes)
	}
	vel, num, vel_err, num_err, activos := InformacionBatches.Tasa()

	log.Println(fmt.Sprintf("Tasa: %.2f Procesados: %d Tasa de Errores: %.2f Numero de Errores: %d Jobs Activos: %d", vel, num, vel_err, num_err, activos))

	//Comprueba si hay actividad
	if activos > 0 {
		//No ha terminado
		paciencia = 5 * 60 / InformacionBatches.Frecuencia
		return false
	}
	//Ha terminado
	if paciencia > 0 {
		return true
	}
	//Parece que ha terminado, pero tengamos paciencia
	paciencia -= 1
	return false
}

package servicio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var InformacionBatches Batches = Batches{Batches: make(map[string]*Batch), Frecuencia: 60}

type Batches struct {
	Batches    map[string]*Batch
	Frecuencia int
}

type Batch struct {
	Procesados_prev int
	Fallados_prev   int
	Pendientes_prev int
	Procesados      int
	Fallados        int
	Pendientes      int
	Acumulado       int
	Acumulado_Err   int
	Activo          bool
}

type Respuesta []struct {
	Batch      string `json:"batch"`
	Procesados string `json:"processed_records"`
	Fallados   string `json:"failed_records"`
	Pendientes string `json:"un_processed_records"`
}

func (batches *Batches) Add(batch string, proc string, fail string, pdte string) {
	val, existe := batches.Batches[batch]
	if existe {
		val.Fallados_prev = val.Fallados
		val.Pendientes_prev = val.Pendientes
		val.Procesados_prev = val.Procesados
		val.Fallados, _ = strconv.Atoi(fail)
		val.Pendientes, _ = strconv.Atoi(pdte)
		val.Procesados, _ = strconv.Atoi(proc)
		val.Acumulado_Err = val.Fallados - val.Fallados_prev
		val.Acumulado = val.Pendientes_prev - val.Pendientes
		if val.Acumulado == 0 {
			val.Activo = false
			log.Println(fmt.Sprintf("El batch %s no tiene actividad", batch))
		} else {
			val.Activo = true
		}
	} else {
		batches.Batches[batch] = creaBatch(proc, fail, pdte)
	}

}

func (batches *Batches) Tasa() (float32, int, float32, int) {
	var total, total_err int
	for _, val := range batches.Batches {
		total += val.Acumulado
		total_err += val.Acumulado_Err
	}
	return float32(total) / float32(batches.Frecuencia) * 60, total, float32(total_err) / float32(batches.Frecuencia) * 60, total_err
}

func creaBatch(proc string, fail string, pdte string) *Batch {

	procesado, _ := strconv.Atoi(proc)
	fallado, _ := strconv.Atoi(fail)
	pendiente, _ := strconv.Atoi(pdte)
	return &Batch{0, 0, 0, procesado, fallado, pendiente, 0, 0, true}
}

func (batches *Batches) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(batches)
}

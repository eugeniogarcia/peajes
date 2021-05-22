package servicio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Batches struct {
	Batches    map[string]*Batch
	Frecuencia int
	Totales    *prometheus.GaugeVec
	Errores    *prometheus.GaugeVec
	Activos    *prometheus.GaugeVec
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
	if !existe {
		val = &Batch{0, 0, 0, 0, 0, 0, 0, 0, true}
		batches.Batches[batch] = val
	}
	//Prepara la medida
	fallados, _ := strconv.Atoi(strings.TrimSpace(fail))
	pendientes, _ := strconv.Atoi(strings.TrimSpace(pdte))
	procesados, _ := strconv.Atoi(strings.TrimSpace(proc))
	//Guarda el valor anterior
	val.Fallados_prev = val.Fallados
	val.Pendientes_prev = val.Pendientes
	val.Procesados_prev = val.Procesados
	val.Fallados = fallados
	val.Pendientes = pendientes
	val.Procesados = procesados
	// Delta entre los dos valores
	val.Acumulado_Err = val.Fallados - val.Fallados_prev
	if val.Pendientes_prev > 0 {
		val.Acumulado = val.Pendientes_prev - val.Pendientes
		if val.Acumulado == 0 {
			val.Activo = false
			log.Println(fmt.Sprintf("El batch %s no tiene actividad", batch))
		} else {
			val.Activo = true
		}
	} else {
		val.Acumulado = 0
		val.Activo = true
	}

	//Actualiza Prometheus
	if batches.Errores != nil {
		batches.Errores.WithLabelValues(batch).Set(float64(fallados))
	}
	//Actualiza Prometheus
	if batches.Totales != nil {
		batches.Totales.WithLabelValues(batch).Set(float64(procesados))
	}
}

func (batches *Batches) Tasa() (float32, int, float32, int, int) {
	var total_acumulado, total_err_acumulado, total, total_err, numero_jobs int

	for _, val := range batches.Batches {
		if val.Activo {
			numero_jobs++
		}
		total += val.Procesados
		total_err += val.Fallados
		total_acumulado += val.Acumulado
		total_err_acumulado += val.Acumulado_Err
	}

	if batches.Errores != nil {
		batches.Errores.WithLabelValues("Total").Set(float64(total_err))
	}

	if batches.Totales != nil {
		batches.Totales.WithLabelValues("Total").Set(float64(total))
	}

	if batches.Activos != nil {
		batches.Activos.WithLabelValues("Total").Set(float64(numero_jobs))
	}

	return float32(total_acumulado) / float32(batches.Frecuencia) * 60, total_acumulado, float32(total_err_acumulado) / float32(batches.Frecuencia) * 60, total_err_acumulado, numero_jobs
}

func (batches *Batches) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(batches)
}

type BatchActivo struct {
	Batch         string
	Procesados    int
	Fallados      int
	Pendientes    int
	Acumulado     float32
	Acumulado_Err float32
}
type BatchNoActivo struct {
	Batch      string
	Procesados int
	Fallados   int
	Pendientes int
}
type ResumenBatches struct {
	BatchesActivo   []*BatchActivo
	BatchesNoActivo []*BatchNoActivo
}

func (batches *Batches) Resumen(rw http.ResponseWriter, r *http.Request) {
	respuesta := ResumenBatches{
		BatchesActivo:   make([]*BatchActivo, 0, 0),
		BatchesNoActivo: make([]*BatchNoActivo, 0, 0)}

	for batch, val := range batches.Batches {
		switch val.Activo {
		case true:
			respuesta.BatchesActivo = append(respuesta.BatchesActivo, &BatchActivo{
				Batch:         batch,
				Procesados:    val.Procesados,
				Fallados:      val.Fallados,
				Pendientes:    val.Pendientes,
				Acumulado:     float32(val.Acumulado) / float32(batches.Frecuencia),
				Acumulado_Err: float32(val.Acumulado_Err) / float32(batches.Frecuencia),
			})
		case false:
			respuesta.BatchesNoActivo = append(respuesta.BatchesNoActivo, &BatchNoActivo{
				Batch:      batch,
				Procesados: val.Procesados,
				Fallados:   val.Fallados,
				Pendientes: val.Pendientes,
			})
		}
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(respuesta)
}

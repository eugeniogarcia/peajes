package servicio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Batches struct {
	Batches    map[string]*Batch
	Frecuencia int
	Paciencia  int
	Cadena     InformacionCadenas
	Totales    *prometheus.GaugeVec
	Errores    *prometheus.GaugeVec
	Activos    *prometheus.GaugeVec
}

type InformacionCadenas map[int][]string
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
	Paciencia       int
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
		val = &Batch{0, 0, 0, 0, 0, 0, 0, 0, true, batches.Paciencia}
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
			val.Paciencia--
			if val.Paciencia > 0 {
				val.Activo = true
			} else {
				val.Activo = false
			}
			//log.Println(fmt.Sprintf("El batch %s no tiene actividad", batch))
		} else {
			val.Paciencia = batches.Paciencia
			val.Activo = true
		}
	} else {
		if !existe {
			val.Acumulado = 0
			val.Paciencia = batches.Paciencia
			val.Activo = true
		} else {
			val.Acumulado = 0
			val.Paciencia--
			if val.Paciencia > 0 {
				val.Activo = true
			} else {
				val.Activo = false
			}
		}
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

type LiteBatches struct {
	BatchesNoActivo  []int
	CadenasNoActivas []int
	CadenasMitad     []int
}

func (batches *Batches) preparaRespuestaLite() LiteBatches {
	respuesta := LiteBatches{
		BatchesNoActivo:  make([]int, 0, 0),
		CadenasNoActivas: make([]int, 0, 0),
		CadenasMitad:     make([]int, 0, 0)}

	if batches.Cadena != nil {
		for cadena_batch, lista_cadenas := range batches.Cadena {
			//Cadena no esta activa
			cadena_activa := false
			cadena_activa_amitad := false
			for pos, val := range lista_cadenas {
				if val == "" {
					continue
				}
				elbatch := batches.Batches[val]
				if elbatch == nil {
					fmt.Println("No encontro el batchid en la respuesta")
					continue
				}
				if elbatch.Activo {
					//Cadena esta activa
					cadena_activa = true
					if pos > 1 {
						cadena_activa_amitad = true
					}
					break
				}
			}
			if !cadena_activa {
				respuesta.CadenasNoActivas = append(respuesta.CadenasNoActivas, cadena_batch)
			} else {
				if cadena_activa_amitad {
					respuesta.CadenasMitad = append(respuesta.CadenasMitad, cadena_batch)
				}
			}
		}
	} else {
		for batch, val := range batches.Batches {
			switch val.Activo {
			case false:
				i, _ := strconv.Atoi(batch)
				respuesta.BatchesNoActivo = append(respuesta.BatchesNoActivo, i)
			}
		}
	}

	return respuesta
}

func (batches *Batches) Lite(rw http.ResponseWriter, r *http.Request) {
	respuesta := batches.preparaRespuestaLite()

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(respuesta)
}

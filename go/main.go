package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/eugeniogarcia/peajes/servicio"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var procesadosTotales = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "instalaciones_total",
		Help: "Numero de instalaciones procesadas.",
	},
	[]string{"batchid"},
)
var erroresTotales = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "instalaciones_error_total",
		Help: "Numero de instalaciones con error.",
	},
	[]string{"batchid"},
)
var activos = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "Jobs_activos",
		Help: "Número de jobs activos.",
	},
	[]string{"batchid"},
)

var peticionesTotales = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Numero de http get requests.",
	},
	[]string{"path"},
)

var estadoRespuesta = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var duracionHttp = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

func init() {
	prometheus.Register(peticionesTotales)
	prometheus.Register(estadoRespuesta)
	prometheus.Register(duracionHttp)
	prometheus.Register(erroresTotales)
	prometheus.Register(procesadosTotales)
	prometheus.Register(activos)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(duracionHttp.WithLabelValues(path))
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		estadoRespuesta.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		peticionesTotales.WithLabelValues(path).Inc()

		timer.ObserveDuration()
	})
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	//Carga la configuración
	var cfg Config
	cargar(&cfg)

	//Prepara los routers
	router := mux.NewRouter()
	router.Use(prometheusMiddleware)
	router.Use(commonMiddleware)

	// Configura el endpoint de Prometheus
	router.Path("/metrics").Handler(promhttp.Handler())
	router.Path("/batches").Handler(&servicio.InformacionBatches)
	router.HandleFunc("/resumen", servicio.InformacionBatches.Resumen)
	router.HandleFunc("/lite", servicio.InformacionBatches.Lite)

	// Indica desde donde poder servir recursos estáticos
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//Prepara la llamada a ISU
	var wg sync.WaitGroup
	wg.Add(1)
	servicio.New(&wg).Start(cfg.ListaBatchs, cfg.Server.Host, cfg.Server.Port, cfg.Frecuencia, cfg.ListaCadenas, procesadosTotales, erroresTotales, activos)

	//Arranca el servidor http
	go func() {
		fmt.Printf("Sirviendo peticiones en el puerto %d", cfg.Puerto)
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Puerto), router)
		log.Fatal(err)
	}()

	//Esperamos hasta que termine el monitoreo de ISU
	wg.Wait()
}

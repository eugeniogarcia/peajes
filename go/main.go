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

	// Configura el endpoint de Prometheus
	router.Path("/metrics").Handler(promhttp.Handler())
	//EGSM
	//router.Path("/batches").Handler()

	// Indica desde donde poder servir recursos estáticos
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//Prepara la llamada a ISU
	var wg sync.WaitGroup
	wg.Add(1)
	servicio.New(&wg).Start(cfg.ListaBatchs, cfg.Server.Host, cfg.Server.Port, cfg.Frecuencia)

	//Arranca el servidor http
	go func() {
		fmt.Println("Sirviendo peticiones en el puerto 9000")
		err := http.ListenAndServe(":9000", router)
		log.Fatal(err)
	}()

	//Esperamos hasta que termine el monitoreo de ISU
	wg.Wait()
}

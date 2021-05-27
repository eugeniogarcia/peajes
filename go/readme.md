# Configure Prometheus

Con `prometheus.yaml` configuramos prometheus para que haga el scraping de metricas.

# Configurar grafana

- Configurar la datasource de Prometheus
- Incluir el dashboard
- Configurar grafana. Incluimos la configuración de la sección _[smtp]_ del archivo _defaults.ini_

# Prometheus

We declare the metrics:

```go
var procesadosTotales = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "instalaciones_total",
		Help: "Numero de instalaciones procesadas.",
	},
	[]string{"batchid"},
)
```

Here we have declared a _Gauge_, and just a label for the metric. This label will be completed by a couple more automatically, _instance_ and _job_. Here is what we may be looking at:

```pql
instalaciones_total{batchid="Total", instance="localhost:9000", job="golang"}
```

The metric has to be registered with the prometheus:

```go
func init() {
	prometheus.Register(procesadosTotales)

```

Finally we have to add to the metrics the data. For example, for this gauge, we may be doing something like this:

```go
batches.Errores.WithLabelValues("Total").Set(float64(total_err))
```

There are other methods available, such as _add_.

## Queries

A couple of examples of prometheus queries:

```pql
rate(instalaciones_total{batchid="Total", instance="localhost:9000", job="golang"}[6m])*60

instalaciones_total{batchid="Total", instance="localhost:9000", job="golang"}
```
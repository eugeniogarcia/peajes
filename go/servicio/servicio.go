package servicio

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"log"
)

var ErrInterrupt = errors.New("Se ha interrumpido la ejecución")

type Runner struct {
	interrupt chan os.Signal
	complete chan error
	wg *sync.WaitGroup
}

func New(wg *sync.WaitGroup) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		wg:wg,
	}
}

func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}

func (r *Runner) Start() {
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		defer r.wg.Done()
		defer signal.Stop(r.interrupt)
		r.run()
	}()
}

func (r *Runner) run() error {
	for  {
		if r.gotInterrupt() {
			log.Println(ErrInterrupt)
			return ErrInterrupt
		}
		//Llama cada x segundos
		//Procesa la respuesta
		//Actualiza las metricas
	}

	return nil
}


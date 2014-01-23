package main

import (
	"encoding/json"
	"fmt"
	metrics "github.com/rcrowley/go-metrics"
	"net"
	"net/http"
	"os"
	"time"
)

var addrs = []*Addr{}

type Addr struct {
	Address  string
	Registry *metrics.Registry
}

var registries = map[string]*metrics.Registry{}

func (a *Addr) startPing() {
	go func(a *Addr) {
		for {
			a.ping()
			time.Sleep(1 * time.Second)
		}
	}(a)
}

func (a *Addr) ping() {
	errors := int64(0)
	start := time.Now()
	conn, err := net.Dial("tcp", a.Address)
	if err != nil {
		errors = 100
	} else {
		conn.Close()
		metrics.GetOrRegisterHistogram(
			"ns",
			*a.Registry,
			metrics.NewUniformSample(60),
		).Update(time.Since(start).Nanoseconds())
	}

	metrics.GetOrRegisterHistogram(
		"errors",
		*a.Registry,
		metrics.NewUniformSample(10),
	).Update(errors)
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	output := map[string]interface{}{}

	for _, addr := range addrs {
		marshal_tmp := map[string]interface{}{}
		// this seems to be the only way to do this
		// FIXME error checking
		content, err := (*addr.Registry).(*metrics.StandardRegistry).MarshalJSON()

		if err != nil {
			w.WriteHeader(500)
			return
		}

		err = json.Unmarshal(content, &marshal_tmp)

		if err != nil {
			w.WriteHeader(500)
			return
		}

		output[addr.Address] = marshal_tmp
	}

	content, _ := json.Marshal(output)
	w.Write(content)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide one or more IP:Port pairs")
		os.Exit(1)
	}

	for _, addr := range os.Args[1:] {
		r := metrics.NewRegistry()
		a := &Addr{
			Address:  addr,
			Registry: &r,
		}

		addrs = append(addrs, a)
		a.startPing()
	}

	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe("127.0.0.1:9117", nil)
}

package httpmetrics

import (
	"encoding/json"
	"github.com/gollector/gollector-monitors/src/util"
	metrics "github.com/rcrowley/go-metrics"
	"net/http"
)

type Handler struct {
	Socket     string
	Registries map[string]*metrics.Registry
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	output := map[string]interface{}{}

	for ip, registry := range h.Registries {
		marshal_tmp := map[string]interface{}{}
		content, err := (*registry).(*metrics.StandardRegistry).MarshalJSON()

		if err != nil {
			w.WriteHeader(500)
			return
		}

		err = json.Unmarshal(content, &marshal_tmp)

		if err != nil {
			w.WriteHeader(500)
			return
		}

		output[ip] = marshal_tmp
	}

	content, _ := json.Marshal(output)
	w.Write(content)
}

func (h *Handler) CreateServer() error {
	s := http.Server{Handler: h}

	l, err := util.CreateSocket(h.Socket)

	if err != nil {
		return err
	}

	return s.Serve(l)
}

package handlers

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"sort"
)

type IndexHandler struct {
	storage MetricStorage
	tmpl    *template.Template
}

type data struct {
	Counters []pair[int64]
	Gauges   []pair[float64]
}

func NewIndexHandler(storage MetricStorage) (*IndexHandler, error) {
	h := IndexHandler{storage: storage}
	err := h.parse()
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh := h.storage.GetAll()

	sortedCounters := mapToSortedSlice(sh.Counters)
	sortedGauges := mapToSortedSlice(sh.Gauges)

	d := data{
		Counters: sortedCounters,
		Gauges:   sortedGauges,
	}

	var buf bytes.Buffer

	err := h.tmpl.ExecuteTemplate(&buf, "index.html", d)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

//go:embed templates
var templates embed.FS

func (h *IndexHandler) parse() error {
	tmpl, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		return err
	}

	h.tmpl = tmpl

	return nil
}

type pair[T any] struct {
	Key   string
	Value T
}

func mapToSortedSlice[T any](m map[string]T) []pair[T] {
	pairs := make([]pair[T], 0, len(m))

	for k, v := range m {
		pairs = append(pairs, pair[T]{Key: k, Value: v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})

	return pairs
}

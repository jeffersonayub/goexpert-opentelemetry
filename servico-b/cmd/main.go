package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jeffersonayub/goexpert-opentelemetry/servico-b/internal/entity"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	http.HandleFunc("GET /{cep}", Handle)
	http.ListenAndServe(":8081", nil)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")
	if !entity.IsValidCEP(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	location, erro, err := entity.GetCep(cep)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	if erro {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	wheater, err := entity.GetWeather(location)
	if err != nil {
		http.Error(w, "can not find weather", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(wheater); err != nil {
		http.Error(w, "can not encode response", http.StatusInternalServerError)
		return
	}
}

func init() {
	endpoint := "http://zipkin:9411/api/v2/spans"
	exporter, err := zipkin.New(endpoint)
	if err != nil {
		log.Fatalf("failed to create zipkin exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("servico-b"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

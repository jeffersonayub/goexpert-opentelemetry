package main

import (
	"encoding/json"
	"net/http"

	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Request struct {
	Cep string `json:"cep"`
}

type Response struct {
	City   string  `json:"city"`
	Temp_C float64 `json:"temp_C"`
	Temp_F float64 `json:"temp_F"`
	Temp_K float64 `json:"temp_K"`
}

func main() {
	http.HandleFunc("POST /", Handle)
	http.ListenAndServe(":8080", nil)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !IsValidCEP(request.Cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, span := otel.Tracer("servico-a").Start(r.Context(), "call-to-servico-b")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://service_b:8081/"+request.Cep, nil)
	if err != nil {
		http.Error(w, "can not create request", http.StatusInternalServerError)
		return
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	var result Response
	json.NewDecoder(response.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "can not encode response", http.StatusInternalServerError)
		return
	}
}

func IsValidCEP(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
			semconv.ServiceNameKey.String("servico-a"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

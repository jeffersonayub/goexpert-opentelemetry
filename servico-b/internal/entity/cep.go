package entity

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Cep struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro,omitempty"`
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

func GetCep(cep string) (localidade string, erro bool, err error) {

	ctx, span := otel.Tracer("servico-b").Start(context.Background(), "call-to-viacep")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	responseCep, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		return "", true, err
	}

	response, err := http.DefaultClient.Do(responseCep)
	if err != nil {
		return "", true, err
	}
	defer response.Body.Close()

	var result Cep
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return "", true, err
	}

	return result.Localidade, result.Erro, nil
}

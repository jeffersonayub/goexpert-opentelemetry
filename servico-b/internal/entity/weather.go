package entity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Weather struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		Temp_C float64 `json:"temp_c"`
		Temp_F float64 `json:"temp_f"`
	} `json:"current"`
}

type Response struct {
	City   string  `json:"city"`
	Temp_C float64 `json:"temp_C"`
	Temp_F float64 `json:"temp_F"`
	Temp_K float64 `json:"temp_K"`
}

const weather_key = "WEATHER_API_KEY" // Replace with your actual API key

func GetWeather(location string) (Response, error) {
	var weather Weather
	err := fetchWeatherData(location, &weather)
	if err != nil {
		return Response{}, err
	}
	return weather.ToResponse(), nil
}

func (w Weather) ToResponse() Response {
	return Response{
		City:   w.Location.Name,
		Temp_C: w.Current.Temp_C,
		Temp_F: w.Current.Temp_F,
		Temp_K: w.Current.Temp_C + 273.15,
	}
}
func fetchWeatherData(location string, weather *Weather) error {
	ctx, span := otel.Tracer("servico-b").Start(context.Background(), "call-to-weatherapi")
	defer span.End()

	span.SetAttributes(attribute.String("location", location))

	responseWeather, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weatherapi.com/v1/current.json?key="+weather_key+"&q="+url.QueryEscape(location), nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(responseWeather)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(weather)
}

package repository

import "github.com/miyamo2/qilin/examples/weather-mcp/domain/model"

type Weather interface {
	GetByCity(city string) (*model.CityWeather, error)
	All() ([]model.CityWeather, error)
}

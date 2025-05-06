package api

import (
	"fmt"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/model"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/repository"
	"math/rand/v2"
	"sync"
	"time"
)

var cities sync.Map

var _ repository.Weather = (*Weather)(nil)

type Weather struct{}

func (w Weather) All() ([]model.CityWeather, error) {
	var cityWeatherList []model.CityWeather
	cities.Range(func(key, value interface{}) bool {
		if cityWeather, ok := value.(model.CityWeather); ok {
			cityWeatherList = append(cityWeatherList, cityWeather)
		}
		return true
	})
	return cityWeatherList, nil
}

func (w Weather) GetByCity(city string) (*model.CityWeather, error) {
	if value, ok := cities.Load(city); ok {
		if cityWeather, ok := value.(model.CityWeather); ok {
			return &cityWeather, nil
		}
	}
	return nil, fmt.Errorf("city '%s' not found", city)
}

func init() {
	cities.Store("tokyo", model.CityWeather{
		City:        "Tokyo",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    65.0,
		Condition:   "sunny",
		WindSpeed:   3.2,
	})
	cities.Store("new_york", model.CityWeather{
		City:        "New York",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    70.0,
		Condition:   "cloudy",
		WindSpeed:   5.1,
	})
	cities.Store("london", model.CityWeather{
		City:        "London",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    75.0,
		Condition:   "rainy",
		WindSpeed:   4.0,
	})

	go func() {
		t1 := time.NewTicker(45 * time.Second)
		defer t1.Stop()

		t2 := time.NewTicker(1*time.Minute + 30*time.Second)
		defer t2.Stop()

		for {
			select {
			case <-time.After(1 * time.Minute):
				cities.Store("paris", model.CityWeather{
					City:        "Paris",
					Date:        time.Now(),
					Temperature: float64(rand.N[int](30)),
					Humidity:    60.0,
					Condition:   "sunny",
				})
			case <-t1.C:
				cities.Store("tokyo", model.CityWeather{
					City:        "Tokyo",
					Date:        time.Now(),
					Temperature: float64(rand.N[int](30)),
					Humidity:    65.0,
					Condition:   "sunny",
					WindSpeed:   3.2,
				})
			case <-t2.C:
				cities.Delete("paris")
			default:
				continue
			}
		}
	}()
}

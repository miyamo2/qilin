package api

import (
	"fmt"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/model"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/repository"
	"math/rand/v2"
	"sync"
	"time"
)

var cities = map[string]model.CityWeather{
	"tokyo": {
		City:        "Tokyo",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    65.0,
		Condition:   "sunny",
		WindSpeed:   3.2,
	},
	"new_york": {
		City:        "New York",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    70.0,
		Condition:   "cloudy",
		WindSpeed:   5.1,
	},
	"london": {
		City:        "London",
		Date:        time.Now(),
		Temperature: float64(rand.N[int](30)),
		Humidity:    75.0,
		Condition:   "rainy",
		WindSpeed:   4.0,
	},
}

var citiesMutex = &sync.RWMutex{}

var _ repository.Weather = (*Weather)(nil)

type Weather struct{}

func (w Weather) All() ([]model.CityWeather, error) {
	var cityWeatherList []model.CityWeather
	for _, weather := range allCityWeather() {
		cityWeatherList = append(cityWeatherList, weather)
	}
	return cityWeatherList, nil
}

func (w Weather) GetByCity(city string) (*model.CityWeather, error) {
	if value, ok := getCityWeather(city); ok {
		return &value, nil
	}
	return nil, fmt.Errorf("city '%s' not found", city)
}

func init() {
	go func() {
		t1 := time.NewTicker(3 * time.Minute)
		defer t1.Stop()

		t2 := time.NewTicker(3*time.Minute + 30*time.Second)
		defer t2.Stop()

		for {
			select {
			case <-time.After(2 * time.Minute):
				storeCityWeather("paris", model.CityWeather{
					City:        "Paris",
					Date:        time.Now(),
					Temperature: float64(rand.N[int](30)),
					Humidity:    60.0,
					Condition:   "sunny",
				})
			case <-t1.C:
				storeCityWeather("tokyo", model.CityWeather{
					City:        "Tokyo",
					Date:        time.Now(),
					Temperature: float64(rand.N[int](30)),
					Humidity:    65.0,
					Condition:   "sunny",
					WindSpeed:   3.2,
				})
			case <-t2.C:
				deleteCityWeather("paris")
			default:
				continue
			}
		}
	}()
}

func storeCityWeather(city string, weather model.CityWeather) {
	citiesMutex.Lock()
	defer citiesMutex.Unlock()
	cities[city] = weather
}

func getCityWeather(city string) (model.CityWeather, bool) {
	citiesMutex.RLock()
	defer citiesMutex.RUnlock()
	weather, exists := cities[city]
	return weather, exists
}

func deleteCityWeather(city string) {
	citiesMutex.Lock()
	defer citiesMutex.Unlock()
	delete(cities, city)
}

func allCityWeather() []model.CityWeather {
	citiesMutex.RLock()
	defer citiesMutex.RUnlock()
	var cityWeatherList []model.CityWeather
	for _, weather := range cities {
		cityWeatherList = append(cityWeatherList, weather)
	}
	return cityWeatherList
}

package model

import "time"

type CityWeather struct {
	City        string    `json:"city"`
	Date        time.Time `json:"date"`
	Temperature float64   `json:"temperature"` // Celsius
	Humidity    float64   `json:"humidity"`    // Relative humidity (%)
	Condition   string    `json:"condition"`   // Weather condition (sunny, cloudy, rainy, etc.)
	WindSpeed   float64   `json:"windSpeed"`   // Wind speed (m/s)
}

func (*CityWeather) Equals(v CityWeather) bool {
	return v.City == v.City &&
		v.Date.Equal(v.Date) &&
		v.Temperature == v.Temperature &&
		v.Humidity == v.Humidity &&
		v.Condition == v.Condition &&
		v.WindSpeed == v.WindSpeed
}

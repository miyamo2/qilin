package handler

import (
	"fmt"
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/model"
	"github.com/miyamo2/qilin/examples/weather-mcp/domain/repository"
	"github.com/miyamo2/qilin/examples/weather-mcp/infrastructure/api"
	"math"
	"math/rand/v2"
	"net/url"
	"strings"
	"time"
)

// ConvertTemperatureFromUnitType represents possible values for from_unit
type ConvertTemperatureFromUnitType string

const (
	ConvertTemperatureFromUnitTypeCelsius    ConvertTemperatureFromUnitType = "celsius"
	ConvertTemperatureFromUnitTypeFahrenheit ConvertTemperatureFromUnitType = "fahrenheit"
)

// ConvertTemperatureToUnitType represents possible values for to_unit
type ConvertTemperatureToUnitType string

const (
	ConvertTemperatureToUnitTypeCelsius    ConvertTemperatureToUnitType = "celsius"
	ConvertTemperatureToUnitTypeFahrenheit ConvertTemperatureToUnitType = "fahrenheit"
)

// ToolConvertTemperatureRequest contains input parameters for the convert_temperature tool.
type ToolConvertTemperatureRequest struct {
	Temperature float64                        `json:"temperature"`
	FromUnit    ConvertTemperatureFromUnitType `json:"from_unit"`
	ToUnit      ConvertTemperatureToUnitType   `json:"to_unit"`
}

// ToolCalculateHumidityIndexRequest contains input parameters for the calculate_humidity_index tool.
type ToolCalculateHumidityIndexRequest struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// CityWeather represents weather data for a city
type CityWeather = model.CityWeather

var repo repository.Weather = &api.Weather{}

func ConvertTemperature(c qilin.ToolContext) error {
	var req ToolConvertTemperatureRequest
	c.Bind(&req)

	temperature := req.Temperature
	fromUnit := req.FromUnit
	toUnit := req.ToUnit

	var result float64

	switch {
	case fromUnit == ConvertTemperatureFromUnitTypeCelsius && toUnit == ConvertTemperatureToUnitTypeFahrenheit:
		// C to F: (C * 9/5) + 32
		result = (temperature * 9 / 5) + 32
	case fromUnit == ConvertTemperatureFromUnitTypeFahrenheit && toUnit == ConvertTemperatureToUnitTypeCelsius:
		// F to C: (F - 32) * 5/9
		result = (temperature - 32) * 5 / 9
	case fromUnit == ConvertTemperatureFromUnitTypeCelsius && toUnit == ConvertTemperatureToUnitTypeCelsius:
		result = temperature
	case fromUnit == ConvertTemperatureFromUnitTypeFahrenheit && toUnit == ConvertTemperatureToUnitTypeFahrenheit:
		result = temperature
	default:
		return fmt.Errorf("unsupported conversion: %s to %s", fromUnit, toUnit)
	}
	// Round to 2 decimal places
	result = math.Round(result*100) / 100
	return c.String(fmt.Sprintf("%.2f %s = %.2f %s", temperature, fromUnit, result, toUnit))
}

func CalculateHumidityIndex(c qilin.ToolContext) error {
	var req ToolCalculateHumidityIndexRequest
	c.Bind(&req)

	temperature := req.Temperature
	humidity := req.Humidity

	// Discomfort index simplified formula: 0.81 × temperature + 0.01 × humidity × (0.99 × temperature - 14.3) + 46.3
	index := 0.81*temperature + 0.01*humidity*(0.99*temperature-14.3) + 46.3
	// Round to 1 decimal place
	index = math.Round(index*10) / 10

	var comfort string
	switch {
	case index < 55:
		comfort = "Cold"
	case index < 60:
		comfort = "Slightly cool"
	case index < 65:
		comfort = "Comfortable"
	case index < 70:
		comfort = "Pleasant"
	case index < 75:
		comfort = "Slightly warm"
	case index < 80:
		comfort = "Warm"
	case index < 85:
		comfort = "Hot"
	default:
		comfort = "Very hot"
	}

	return c.String(fmt.Sprintf("Temperature: %.1f°C, Humidity: %.1f%%\nComfort Index: %.1f (%s)", temperature, humidity, index, comfort))
}

func GetWeatherForecast(c qilin.ResourceContext) error {
	city := c.Param("city")
	if city == "" {
		return fmt.Errorf("city is required")
	}

	weather, err := repo.GetByCity(city)
	if err != nil {
		return err
	}

	return c.JSON(weather)
}

func ResourceList(c qilin.ResourceListContext) error {
	cityWeathers, err := repo.All()
	if err != nil {
		return err
	}
	for _, v := range cityWeathers {
		city := strings.ReplaceAll(strings.ToLower(v.City), " ", "_")
		uri, err := url.Parse(fmt.Sprintf("weather://forecast/%s", city))
		if err != nil {
			return err
		}
		c.SetResource(uri.String(), qilin.Resource{
			URI:         (*qilin.ResourceURI)(uri),
			Name:        fmt.Sprintf("%s Weather Forecast", v.City),
			Description: fmt.Sprintf("Current weather data for %s", v.City),
			MimeType:    "application/json",
		})
	}
	return nil
}

func WeatherReport(c qilin.PromptContext) error {
	cityName := c.Param("city")
	city, err := repo.GetByCity(cityName)
	if err != nil {
		return err
	}

	c.String(qilin.PromptRoleUser, fmt.Sprintf("Please provide a weather report for %s", city.City))

	language := "en"
	if v := c.Param("language"); v != "" {
		language = v
	}

	if language == "ja" {
		return c.String(qilin.PromptRoleUser, fmt.Sprintf(
			"%sの天気レポートです。現在の気温は%.1f℃、湿度は%.1f%%、天候は%sで、風速は%.1fm/sです。",
			city.City, city.Temperature, city.Humidity, translateCondition(city.Condition, "ja"), city.WindSpeed,
		))
	}
	return c.String(qilin.PromptRoleUser, fmt.Sprintf(
		"Weather report for %s. Current temperature is %.1f°C, humidity is %.1f%%, weather condition is %s, and wind speed is %.1f m/s.",
		city.City, city.Temperature, city.Humidity, city.Condition, city.WindSpeed,
	))
}

func WeatherAlert(c qilin.PromptContext) error {
	alertType := c.Param("alert_type")
	severity := c.Param("severity")

	c.String(
		qilin.PromptRoleUser,
		fmt.Sprintf("Generate a weather alert for %s with severity %s", alertType, severity))
	switch alertType {
	case "rain":
		return c.String(
			qilin.PromptRoleAssistant,
			fmt.Sprintf("WEATHER ALERT: Heavy rain warning. Severity level: %s/5. Expect heavy rainfall and possible flooding in low-lying areas. Please take necessary precautions.", severity))
	case "heat":
		return c.String(
			qilin.PromptRoleAssistant,
			fmt.Sprintf("WEATHER ALERT: Heat warning. Severity level: %s/5. Extremely high temperatures expected. Stay hydrated and avoid direct sun exposure.", severity))
	default:
		return c.String(
			qilin.PromptRoleAssistant,
			fmt.Sprintf("WEATHER ALERT: %s warning. Severity level: %s/5. Please stay informed about changing weather conditions.", alertType, severity))
	}
}

func WeatherForecastChangeObserver(c qilin.ResourceChangeContext) {
	for t := range time.Tick(time.Minute) {
		select {
		case <-c.Context().Done():
			return
		default:
			// Assume that a random resource has been changed and send a notification.
			cityWeathers, err := repo.All()
			if err != nil {
				continue
			}
			idx := rand.N[int](len(cityWeathers))
			cityWeather := cityWeathers[idx]
			city := strings.ReplaceAll(strings.ToLower(cityWeather.City), " ", "_")
			uri, _ := url.Parse(fmt.Sprintf("weather://forecast/%s", city))
			c.Publish(uri, t)
		}
	}
}

func ResourceListChangeObserver(c qilin.ResourceListChangeContext) {
	for t := range time.Tick(2 * time.Minute) {
		select {
		case <-c.Context().Done():
			return
		default:
			// Assume that a resource list has been changed and send a notification.
			c.Publish(t)
		}
	}
}

// translateCondition translates weather conditions to the specified language
func translateCondition(condition, language string) string {
	if language != "ja" {
		return condition
	}

	translations := map[string]string{
		"sunny":    "晴れ",
		"cloudy":   "曇り",
		"rainy":    "雨",
		"snowy":    "雪",
		"foggy":    "霧",
		"windy":    "強風",
		"stormy":   "嵐",
		"thunder":  "雷",
		"drizzle":  "小雨",
		"blizzard": "吹雪",
	}

	if translation, ok := translations[condition]; ok {
		return translation
	}
	return condition
}

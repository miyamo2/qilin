package main

import (
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/examples/weather-mcp/handler"
)

func main() {
	q := qilin.New("weather")

	q.Tool("convert_temperature",
		"Convert temperature between Celsius and Fahrenheit",
		(*handler.ToolConvertTemperatureRequest)(nil),
		handler.ConvertTemperature)

	q.Tool("calculate_humidity_index",
		"Calculate humidity index based on temperature and humidity",
		(*handler.ToolCalculateHumidityIndexRequest)(nil),
		handler.CalculateHumidityIndex)

	q.Resource(
		"City Weather Forecast",
		"weather://forecast/{city}",
		"Weather forecast for a specific city",
		handler.GetWeatherForecast,
		qilin.WithResourceMimeType("text/plain"))

	q.ResourceChangeObserver("weather://forecast/{city}", handler.WeatherForecastChangeObserver)

	q.ResourceList(handler.ResourceList)

	q.ResourceListChangeObserver(handler.ResourceListChangeObserver)

	q.Start()
}

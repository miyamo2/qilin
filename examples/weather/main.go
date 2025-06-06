package main

import (
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/examples/weather-mcp/handler"
)

func main() {
	q := qilin.New("weather")

	q.Tool("convert_temperature",
		(*handler.ToolConvertTemperatureRequest)(nil),
		handler.ConvertTemperature,
		qilin.ToolWithDescription("Convert temperature between Celsius and Fahrenheit"))

	q.Tool("calculate_humidity_index",
		(*handler.ToolCalculateHumidityIndexRequest)(nil),
		handler.CalculateHumidityIndex,
		qilin.ToolWithDescription("Calculate humidity index based on temperature and humidity"))

	q.Resource(
		"City Weather Forecast",
		"weather://forecast/{city}",
		handler.GetWeatherForecast,
		qilin.ResourceWithDescription("Weather forecast for a specific city"),
		qilin.ResourceWithMimeType("text/plain"))

	q.ResourceChangeObserver("weather://forecast/{city}", handler.WeatherForecastChangeObserver)

	q.ResourceList(handler.ResourceList)

	q.ResourceListChangeObserver(handler.ResourceListChangeObserver)

	q.Start()
}

package main

import (
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/examples/weather-mcp/handler"
	"github.com/miyamo2/qilin/transport"
	"net"
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

	q.Prompt("weather_report", handler.WeatherReport,
		qilin.PromptWithDescription("Generate a weather report based on weather data"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "city",
				Description: "City name",
				Required:    true,
			}, qilin.PromptArgument{
				Name:        "language",
				Description: "Report language (e.g. 'en', 'ja')",
			}))

	q.Prompt("weather_alert", handler.WeatherAlert,
		qilin.PromptWithDescription("Weather alert"),
		qilin.PromptWithArguments(
			qilin.PromptArgument{
				Name:        "alert_type",
				Description: "Type of alert (e.g. 'rain', 'snow', 'heat')",
				Required:    true,
			},
			qilin.PromptArgument{
				Name:        "severity",
				Description: "Alert severity (1-5)",
				Required:    true,
			}))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	q.Start(qilin.StartWithListener(transport.NewStreamable(transport.StreamableWithNetListener(listener))))
}

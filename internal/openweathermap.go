// Package internal demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current weather for the cities
package internal

import (
	"fmt"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/hspaay/iotc.golang/publisher"
)

// CurrentWeatherInst instance name for current weather
var CurrentWeatherInst = "current"

// LastHourWeatherInst instance name for last 1 hour weather (eg rain, snow)
var LastHourWeatherInst = "hour"

// ForecastWeatherInst instance name for upcoming forecast
var ForecastWeatherInst = "forecast"

// AppID default value. Can be overridden in config.
const AppID = "openweathermap"

// KelvinToC is nr of Kelvins at 0 degrees. openweathermap reports temp in Kelvin
// const KelvinToC = 273.1 // Kelvin at 0 celcius

// var weatherPub *publisher.PublisherState

// WeatherApp with application state, loaded from openweathermap.conf
type WeatherApp struct {
	Cities      []string `yaml:"cities"`
	APIKey      string   `yaml:"apikey"`
	PublisherID string   `yaml:"publisherId"`
}

// PublishNodes creates the nodes and outputs
func (weatherApp *WeatherApp) PublishNodes(pub *publisher.Publisher) {
	// pubNode := weatherPub.PublisherNode()
	// outputs := pub.Outputs

	// Create a node for each city with temperature outputs
	for _, city := range weatherApp.Cities {
		pub.NewNode(city, iotc.NodeTypeWeatherService)

		lc := nodes.NewNodeConfig("language", iotc.DataTypeEnum, "Reporting language. See https://openweathermap.org/current for more options", "en")
		pub.Nodes.UpdateNodeConfig(city, lc)

		// Add individual outputs for each weather info type
		pub.NewOutput(city, iotc.OutputTypeWeather, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeTemperature, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeHumidity, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeAtmosphericPressure, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeWindHeading, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeWindSpeed, CurrentWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeRain, LastHourWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeSnow, LastHourWeatherInst)

		// todo: Add outputs for various forecasts. This needs a paid account so maybe some other time.
		pub.NewOutput(city, iotc.OutputTypeWeather, ForecastWeatherInst)
		pub.NewOutput(city, iotc.OutputTypeTemperature, "max")
		pub.NewOutput(city, iotc.OutputTypeAtmosphericPressure, "min")
	}
}

// UpdateWeather obtains the weather and publishes the output value
// node:city -
//             type: weather    - instance: current, message: value
//             type: temperature- instance: current, message: value
//             type: humidity   - instance: current, message: value
//             etc...
// The iotconnect library will automatically publish changes to the values
func (weatherApp *WeatherApp) UpdateWeather(weatherPub *publisher.Publisher) {

	apikey := weatherApp.APIKey
	weatherPub.Logger().Info("UpdateWeather")

	// publish the current weather for each of the city nodes
	for _, node := range weatherPub.Nodes.GetAllNodes() {
		language := node.Attr["language"]
		currentWeather, err := GetCurrentWeather(apikey, node.NodeID, language)
		if err != nil {
			weatherPub.SetNodeErrorStatus(node.NodeID, iotc.NodeRunStateError, "Current weather not available: "+err.Error())
		} else {
			weatherPub.SetNodeErrorStatus(node.NodeID, iotc.NodeRunStateReady, "")

			var weatherDescription string = ""
			if len(currentWeather.Weather) > 0 {
				weatherDescription = currentWeather.Weather[0].Description
			}
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeWeather, CurrentWeatherInst, weatherDescription)
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeTemperature, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Main.Temperature))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeHumidity, CurrentWeatherInst, fmt.Sprintf("%d", currentWeather.Main.Humidity))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeAtmosphericPressure, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Main.Pressure))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeWindSpeed, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Wind.Speed))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeWindHeading, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Wind.Heading))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeRain, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Rain.LastHour*1000))
			weatherPub.UpdateOutputValue(node.NodeID, iotc.OutputTypeSnow, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Snow.LastHour*1000))
		}
	}

	// TODO: move to its own 6 hour interval
	// weatherApp.UpdateForecast(weatherPub)
}

// UpdateForecast obtains a daily forecast and publishes this as a $forecast command
// This is published as follows: zone/publisher/node=city/$forecast/{type}/{instance}
//
// Note this requires a paid account - untested
func (weatherApp *WeatherApp) UpdateForecast(weatherPub *publisher.Publisher) {
	apikey := weatherApp.APIKey

	// publish the daily forecast weather for each of the city nodes
	for _, node := range weatherPub.Nodes.GetAllNodes() {
		language := node.Attr["language"]
		dailyForecast, err := GetDailyForecast(apikey, node.NodeID, language)
		if err != nil {
			weatherPub.SetNodeErrorStatus(node.Address, iotc.NodeRunStateError, "UpdateForecast: Error getting the daily forecast")
			return
		} else if dailyForecast.List == nil {
			weatherPub.SetNodeErrorStatus(node.Address, iotc.NodeRunStateError, "UpdateForecast: Daily forecast not provided")
			return
		}
		weatherPub.SetNodeErrorStatus(node.Address, iotc.NodeRunStateReady, "")

		// build forecast history lists of weather and temperature forecasts
		// TODO: can this be done as a future history publication instead?
		weatherList := make(nodes.OutputForecast, 0)
		maxTempList := make(nodes.OutputForecast, 0)
		minTempList := make(nodes.OutputForecast, 0)

		for _, forecast := range dailyForecast.List {
			epochTime := int64(forecast.Date)
			timestamp := time.Unix(epochTime, 0).Format(iotc.TimeFormat)
			outputValue := iotc.OutputValue{Timestamp: timestamp, EpochTime: epochTime, Value: ""}

			// add the weather descriptions
			var weatherDescription string = ""
			if len(forecast.Weather) > 0 {
				weatherDescription = forecast.Weather[0].Description
			}
			outputValue.Value = weatherDescription
			weatherList = append(weatherList, outputValue)
			outputValue.Value = fmt.Sprintf("%.1f", forecast.Temp.Max)
			maxTempList = append(maxTempList, outputValue)
			outputValue.Value = fmt.Sprintf("%.1f", forecast.Temp.Min)
			minTempList = append(maxTempList, outputValue)
		}
		cityAddress := node.Address
		outputForecasts := weatherPub.OutputForecasts
		outputForecasts.UpdateForecast(cityAddress, iotc.OutputTypeWeather, ForecastWeatherInst, weatherList)
		outputForecasts.UpdateForecast(cityAddress, iotc.OutputTypeTemperature, "max", maxTempList)
		outputForecasts.UpdateForecast(cityAddress, iotc.OutputTypeTemperature, "min", minTempList)
	}
}

// OnNodeConfigHandler handles requests to update node configuration
func (weatherApp *WeatherApp) OnNodeConfigHandler(
	node *iotc.NodeDiscoveryMessage, config iotc.NodeAttrMap) iotc.NodeAttrMap {
	return nil
}

// NewWeatherApp creates the weather app
func NewWeatherApp() *WeatherApp {
	app := WeatherApp{
		Cities:      make([]string, 0),
		PublisherID: AppID,
	}
	return &app
}

// Run the publisher until the SIGTERM  or SIGINT signal is received
func Run() {
	weatherApp := NewWeatherApp()
	weatherPub, _ := publisher.NewAppPublisher("openweathermap", "", "", &weatherApp, true)

	// Discover the node(s) and outputs. Use default for republishing discovery
	weatherPub.SetDiscoveryInterval(0, weatherApp.PublishNodes)
	// Update the forecast once an hour
	weatherPub.SetPollInterval(3600, weatherApp.UpdateWeather)

	// handle update of node configuraiton
	weatherPub.SetNodeConfigHandler(weatherApp.OnNodeConfigHandler)
	// handle update of node inputs
	// weatherPub.SetNodeInputHandler( weatherApp.OnNodeInputHandler)

	// Handle input commands to delete/remove a city node
	// weatherPub.SetNodeCommandHandler(standard.CommandCreate, onAddCity)
	// weatherPub.SetNodeCommandHandler(standard.CommandDelete, onRemoveCity)
	// Handle city node configuration
	// weatherPub.SetNodeCommandHandler(standard.CommandConfig, onConfig)
	// Handle set command (n/a)
	// weatherPub.SetNodeCommandHandler(standard.CommandInput, onInput)

	weatherPub.Start()
	weatherPub.WaitForSignal()
	weatherPub.Stop()
}

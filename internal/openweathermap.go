// Package internal demonstrates building a simple publisher for weather forecasts
// This publishes the current weather for the cities
package internal

import (
	"fmt"
	"time"

	"github.com/iotdomain/iotdomain-go/outputs"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
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

	// Create a node for each city with temperature outputs. The city name is the node ID
	for _, city := range weatherApp.Cities {
		pub.UpdateNodeConfig(city, "language", &types.ConfigAttr{
			DataType:    types.DataTypeEnum,
			Description: "Reporting language. See https://openweathermap.org/current for more options",
			Default:     "en",
		})

		// Add individual outputs for each weather info type
		pub.CreateOutput(city, types.OutputTypeWeather, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeTemperature, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeHumidity, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeAtmosphericPressure, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeWindHeading, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeWindSpeed, CurrentWeatherInst)
		pub.CreateOutput(city, types.OutputTypeRain, LastHourWeatherInst)
		pub.CreateOutput(city, types.OutputTypeSnow, LastHourWeatherInst)

		// todo: Add outputs for various forecasts. This needs a paid account so maybe some other time.
		pub.CreateOutput(city, types.OutputTypeWeather, ForecastWeatherInst)
		pub.CreateOutput(city, types.OutputTypeTemperature, "max")
		pub.CreateOutput(city, types.OutputTypeAtmosphericPressure, "min")
	}
}

// UpdateWeather obtains the weather and publishes the output value
// node:city -
//             type: weather    - instance: current, message: value
//             type: temperature- instance: current, message: value
//             type: humidity   - instance: current, message: value
//             etc...
// The go-iotdomain library will automatically publish changes to the values
func (weatherApp *WeatherApp) UpdateWeather(weatherPub *publisher.Publisher) {

	apikey := weatherApp.APIKey
	logrus.Info("UpdateWeather start")

	weatherApp.PublishNodes(weatherPub)

	// publish the current weather for each of the city nodes
	for _, node := range weatherPub.GetNodes() {
		language := node.Attr["language"]
		startTime := time.Now()
		currentWeather, err := GetCurrentWeather(apikey, node.NodeID, language)
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		if err != nil {
			weatherPub.UpdateNodeErrorStatus(node.NodeID, types.NodeRunStateError, "Current weather not available: "+err.Error())
		} else {
			weatherPub.UpdateNodeStatus(node.NodeID, map[types.NodeStatus]string{
				types.NodeStatusRunState:    string(types.NodeRunStateReady),
				types.NodeStatusLastError:   "",
				types.NodeStatusLatencyMSec: fmt.Sprintf("%d", latency.Milliseconds()),
			})

			var weatherDescription string = ""
			if len(currentWeather.Weather) > 0 {
				weatherDescription = currentWeather.Weather[0].Description
			}
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeWeather, CurrentWeatherInst, weatherDescription)
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeTemperature, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Main.Temperature))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeHumidity, CurrentWeatherInst, fmt.Sprintf("%d", currentWeather.Main.Humidity))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeAtmosphericPressure, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Main.Pressure))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeWindSpeed, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Wind.Speed))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeWindHeading, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Wind.Heading))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeRain, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Rain.LastHour*1000))
			weatherPub.UpdateOutputValue(node.NodeID, types.OutputTypeSnow, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Snow.LastHour*1000))
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
	for _, node := range weatherPub.GetNodes() {
		language := node.Attr["language"]
		dailyForecast, err := GetDailyForecast(apikey, node.NodeID, language)
		if err != nil {
			weatherPub.UpdateNodeErrorStatus(node.Address, types.NodeRunStateError, "UpdateForecast: Error getting the daily forecast")
			return
		} else if dailyForecast.List == nil {
			weatherPub.UpdateNodeErrorStatus(node.Address, types.NodeRunStateError, "UpdateForecast: Daily forecast not provided")
			return
		}
		weatherPub.UpdateNodeErrorStatus(node.Address, types.NodeRunStateReady, "")

		// build forecast history lists of weather and temperature forecasts
		// TODO: can this be done as a future history publication instead?
		weatherList := make(outputs.OutputForecast, 0)
		maxTempList := make(outputs.OutputForecast, 0)
		minTempList := make(outputs.OutputForecast, 0)

		for _, forecast := range dailyForecast.List {
			epochTime := int64(forecast.Date)
			timestamp := time.Unix(epochTime, 0).Format(types.TimeFormat)
			outputValue := types.OutputValue{Timestamp: timestamp, EpochTime: epochTime, Value: ""}

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
		outputID := outputs.MakeOutputID(cityAddress, types.OutputTypeWeather, ForecastWeatherInst)
		weatherPub.UpdateOutputForecast(outputID, weatherList)
		outputID = outputs.MakeOutputID(cityAddress, types.OutputTypeTemperature, "max")
		weatherPub.UpdateOutputForecast(cityAddress, maxTempList)
		outputID = outputs.MakeOutputID(cityAddress, types.OutputTypeTemperature, "min")
		weatherPub.UpdateOutputForecast(cityAddress, minTempList)
	}
}

// OnNodeConfigHandler handles requests to update node configuration
func (weatherApp *WeatherApp) OnNodeConfigHandler(nodeAddress string, config types.NodeAttrMap) types.NodeAttrMap {
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
	weatherPub, _ := publisher.NewAppPublisher("openweathermap", "", &weatherApp, true)

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

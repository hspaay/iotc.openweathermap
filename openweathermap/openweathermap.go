// Package openweathermap demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current weather for the cities
package openweathermap

import (
	"fmt"
	"time"

	"github.com/hspaay/iotc.golang/messaging"
	"github.com/hspaay/iotc.golang/nodes"
	"github.com/hspaay/iotc.golang/publisher"
)

// CurrentWeatherInst instance name for current weather
var CurrentWeatherInst = "current"

// LastHourWeatherInst instance name for last 1 hour weather (eg rain, snow)
var LastHourWeatherInst = "hour"

// ForecastWeatherInst instance name for upcoming forecast
var ForecastWeatherInst = "forecast"

// PublisherID default value. Can be overridden in config.
const PublisherID = "openweathermap"

// KelvinToC is nr of Kelvins at 0 degrees. openweathermap reports temp in Kelvin
// const KelvinToC = 273.1 // Kelvin at 0 celcius

// var weatherPub *publisher.PublisherState

// WeatherApp with application state, loaded from openweathermap.conf
type WeatherApp struct {
	Cities      []string `yaml:"cities"`
	APIKey      string   `yaml:"apikey"`
	PublisherID string   `yaml:"publisher"`
}

// PublishNodes creates the nodes and outputs
func (weatherApp *WeatherApp) PublishNodes(weatherPub *publisher.Publisher) {
	pubNode := weatherPub.PublisherNode
	zone := pubNode.Zone
	outputs := weatherPub.Outputs

	// Create a node for each city with temperature outputs
	for _, city := range weatherApp.Cities {
		cityNode := nodes.NewNode(zone, weatherApp.PublisherID, city, messaging.NodeTypeWeatherForecast)
		weatherPub.Nodes.UpdateNode(cityNode)

		lc := nodes.NewConfigAttr("language", messaging.DataTypeEnum, "Reporting language. See https://openweathermap.org/current for more options", "en")
		weatherPub.Nodes.SetNodeConfig(cityNode.Address, lc)

		// Add individual outputs for each weather info type
		outputs.NewOutput(cityNode, messaging.OutputTypeWeather, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeTemperature, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeHumidity, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeAtmosphericPressure, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeWindHeading, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeWindSpeed, CurrentWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeRain, LastHourWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeSnow, LastHourWeatherInst)

		// todo: Add outputs for various forecasts. This needs a paid account so maybe some other time.
		outputs.NewOutput(cityNode, messaging.OutputTypeWeather, ForecastWeatherInst)
		outputs.NewOutput(cityNode, messaging.OutputTypeTemperature, "max")
		outputs.NewOutput(cityNode, messaging.OutputTypeAtmosphericPressure, "min")
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
	outputHistory := weatherPub.OutputValues
	weatherPub.Logger.Info("UpdateWeather")

	// publish the current weather for each of the city nodes
	for _, node := range weatherPub.Nodes.GetAllNodes() {
		if node.ID != messaging.PublisherNodeID {
			language := node.Config["language"].Value
			currentWeather, err := GetCurrentWeather(apikey, node.ID, language)
			if err != nil {
				weatherPub.Nodes.SetErrorStatus(node, "Current weather not available")
				return
			}
			var weatherDescription string = ""
			if len(currentWeather.Weather) > 0 {
				weatherDescription = currentWeather.Weather[0].Description
			}
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeWeather, CurrentWeatherInst, weatherDescription)
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeTemperature, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Main.Temperature))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeHumidity, CurrentWeatherInst, fmt.Sprintf("%d", currentWeather.Main.Humidity))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeAtmosphericPressure, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Main.Pressure))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeWindSpeed, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Wind.Speed))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeWindHeading, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Wind.Heading))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeRain, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Rain.LastHour*1000))
			outputHistory.UpdateOutputValue(node, messaging.OutputTypeSnow, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Snow.LastHour*1000))
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
		if node.ID != messaging.PublisherNodeID {
			language := node.Config["language"].Value
			dailyForecast, err := GetDailyForecast(apikey, node.ID, language)
			if err != nil {
				weatherPub.Nodes.SetErrorStatus(node, "Error getting the daily forecast")
				return
			} else if dailyForecast.List == nil {
				weatherPub.Nodes.SetErrorStatus(node, "Daily forecast not provided")
				return
			}
			// build forecast history lists of weather and temperature forecasts
			// TODO: can this be done as a future history publication instead?
			weatherList := make(messaging.OutputHistoryList, 0)
			maxTempList := make(messaging.OutputHistoryList, 0)
			minTempList := make(messaging.OutputHistoryList, 0)

			for _, forecast := range dailyForecast.List {
				epochTime := int64(forecast.Date)
				timestamp := time.Unix(epochTime, 0).Format(messaging.TimeFormat)
				outputValue := messaging.OutputValue{Timestamp: timestamp, EpochTime: epochTime, Value: ""}

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
			weatherPub.UpdateForecast(node, messaging.OutputTypeWeather, ForecastWeatherInst, weatherList)
			weatherPub.UpdateForecast(node, messaging.OutputTypeTemperature, "max", maxTempList)
			weatherPub.UpdateForecast(node, messaging.OutputTypeTemperature, "min", minTempList)

		}
	}
}

// OnNodeConfigHandler handles requests to update node configuration
func (weatherApp *WeatherApp) OnNodeConfigHandler(
	node *nodes.Node, config messaging.NodeAttrMap) messaging.NodeAttrMap {
	return nil
}

// NewWeatherApp creates the weather app
func NewWeatherApp() *WeatherApp {
	app := WeatherApp{
		Cities:      make([]string, 0),
		PublisherID: PublisherID,
	}
	return &app
}

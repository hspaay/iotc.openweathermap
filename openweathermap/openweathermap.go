// Package openweathermap demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current weather for the cities
package openweathermap

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
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
func (weatherApp *WeatherApp) PublishNodes(weatherPub *publisher.PublisherState) {
	pubNode := weatherPub.GetNodeByID(standard.PublisherNodeID)

	// Create a node for each city with temperature outputs
	for _, city := range weatherApp.Cities {
		node := weatherPub.DiscoverNode(standard.NewNode(pubNode.Zone, pubNode.PublisherID, city))
		lc := standard.NewConfig("language", standard.DataTypeEnum, "Reporting language. See https://openweathermap.org/current for more options", "en")
		weatherPub.DiscoverNodeConfig(node, lc)

		// Add individual outputs for each weather info type
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeWeather, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeTemperature, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeHumidity, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeAtmosphericPressure, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeWindHeading, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeWindSpeed, CurrentWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeRain, LastHourWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeSnow, LastHourWeatherInst))

		// todo: Add outputs for forecast
		out := standard.NewOutput(node, standard.IOTypeWeather, ForecastWeatherInst)
		out.DataType = standard.DataTypeList
		weatherPub.DiscoverOutput(out)
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeTemperature, ForecastWeatherInst))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeAtmosphericPressure, ForecastWeatherInst))
	}
}

// UpdateWeather obtains the weather and updates the output value
// The iotconnect library will automatically publish changes to the values
func (weatherApp *WeatherApp) UpdateWeather(weatherPub *publisher.PublisherState) {
	// pubNode := weatherPub.GetNodeByID(standard.PublisherNodeID)
	apikey := weatherApp.APIKey

	// publish the current weather for each of the city nodes
	for _, node := range weatherPub.GetAllNodes() {
		if node.ID != standard.PublisherNodeID {
			language := node.Config["language"].Value
			currentWeather, err := GetCurrentWeather(apikey, node.ID, language)
			if err != nil {
				weatherPub.SetErrorStatus(node, "Current weather not available")
				return
			}
			var weatherDescription string = ""
			if len(currentWeather.Weather) > 0 {
				weatherDescription = currentWeather.Weather[0].Description
			}
			weatherPub.UpdateOutputValue(node, standard.IOTypeWeather, CurrentWeatherInst, weatherDescription)
			weatherPub.UpdateOutputValue(node, standard.IOTypeTemperature, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Main.Temperature))
			weatherPub.UpdateOutputValue(node, standard.IOTypeHumidity, CurrentWeatherInst, fmt.Sprintf("%d", currentWeather.Main.Humidity))
			weatherPub.UpdateOutputValue(node, standard.IOTypeAtmosphericPressure, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Main.Pressure))
			weatherPub.UpdateOutputValue(node, standard.IOTypeWindSpeed, CurrentWeatherInst, fmt.Sprintf("%.1f", currentWeather.Wind.Speed))
			weatherPub.UpdateOutputValue(node, standard.IOTypeWindHeading, CurrentWeatherInst, fmt.Sprintf("%.0f", currentWeather.Wind.Heading))
			weatherPub.UpdateOutputValue(node, standard.IOTypeRain, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Rain.LastHour*1000))
			weatherPub.UpdateOutputValue(node, standard.IOTypeSnow, LastHourWeatherInst, fmt.Sprintf("%.1f", currentWeather.Snow.LastHour*1000))
		}
	}

	// TODO: move to its own 6 hour interval
	weatherApp.UpdateForecast(weatherPub)
}

// UpdateForecast obtains the 5 day weather forecast and updates the output value
// The iotconnect library will automatically publish changes to the values
// Update this every 3 to 12 hours
// Note that this isn't a common IoT problem to solve.
func (weatherApp *WeatherApp) UpdateForecast(weatherPub *publisher.PublisherState) {
	apikey := weatherApp.APIKey

	// publish the forecast weather for each of the city nodes
	for _, node := range weatherPub.GetAllNodes() {
		if node.ID != standard.PublisherNodeID {
			language := node.Config["language"].Value
			forecastWeather, err := Get5DayForecast(apikey, node.ID, language)
			if err != nil {
				weatherPub.SetErrorStatus(node, "Forecast not available")
				return
			}
			// build lists of weather and temperature forecasts
			type ForecastValue struct {
				TimeStamp string `json:"timestamp"`
				Value     string `json:"value"`
			}
			weatherList := make([]ForecastValue, 0)
			tempList := make([]ForecastValue, 0)
			if forecastWeather.List != nil {
				// add each forecast to the value lists
				for _, forecast := range forecastWeather.List {
					timestamp := time.Unix(int64(forecast.Date), 0)
					timestampStr := timestamp.Format(time.RFC3339)
					// add the weather descriptions
					var weatherDescription string = ""
					if len(forecast.Weather) > 0 {
						weatherDescription = forecast.Weather[0].Description
					}
					weatherList = append(weatherList, ForecastValue{TimeStamp: timestampStr, Value: weatherDescription})
					// add temp
					temp := fmt.Sprintf("%.1f", forecast.Main.Temperature)
					tempList = append(tempList, ForecastValue{TimeStamp: timestampStr, Value: temp})
					// publish as forecast with multiple values in content, similar to history
					_ = forecast
				}
				weatherListJSON, _ := json.MarshalIndent(weatherList, " ", " ")
				tempListJSON, _ := json.MarshalIndent(tempList, " ", " ")
				weatherPub.UpdateOutputValue(node, standard.IOTypeWeather, ForecastWeatherInst, string(weatherListJSON))
				weatherPub.UpdateOutputValue(node, standard.IOTypeTemperature, ForecastWeatherInst, string(tempListJSON))

			}
			// todo
		}
	}
}

// NewWeatherApp creates the weather app
func NewWeatherApp() *WeatherApp {
	app := WeatherApp{
		Cities:      make([]string, 0),
		PublisherID: PublisherID,
	}
	return &app
}

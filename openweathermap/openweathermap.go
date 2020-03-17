// Package openweathermap demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current temperature and humidity for the cities
package openweathermap

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
)

var ioCurrentInstance = "current"

// PublisherID default value. Can be overridden in config.
const PublisherID = "openweathermap"

// KelvinToC is nr of Kelvins at 0 degrees. openweathermap reports temp in Kelvin
const KelvinToC = 273.1 // Kelvin at 0 celcius

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

		// Add outputs for the current temperature and humidity
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeTemperature, ioCurrentInstance))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeHumidity, ioCurrentInstance))
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
			currentWeather, err := GetCurrentWeather(apikey, node.ID)
			if err != nil {
				weatherPub.SetErrorStatus(node, "Forecast not available")
				return
			}
			weatherPub.UpdateOutputValue(node, standard.IOTypeTemperature, ioCurrentInstance, fmt.Sprintf("%f", currentWeather.Main.Temperature-KelvinToC))
			weatherPub.UpdateOutputValue(node, standard.IOTypeHumidity, ioCurrentInstance, fmt.Sprintf("%d", currentWeather.Main.Humidity))
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

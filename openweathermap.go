// Package openweathermap demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current temperature and humidity for the cities
package openweathermap

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/config"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/sirupsen/logrus"
)

var ioCurrentInstance = "current"

// default config folder
const publisherID = "openweathermap"
const KelvinToC = 273.1 // Kelvin at 0 celcius

// var weatherPub *publisher.PublisherState

// WeatherApp with application state, loaded from openweathermap.conf
type WeatherApp struct {
	Cities []string `yaml:"cities"`
	APIKey string   `yaml:"apikey"`
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

func main() {
	logger := logrus.New()

	messengerConfig := &messenger.MessengerConfig{}
	weatherApp := WeatherApp{}
	config.LoadAppConfig("", publisherID, messengerConfig, &weatherApp)

	messenger := messenger.NewMqttMessenger(messengerConfig, logger)
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, publisherID, messenger)

	// Discover the node(s) and outputs. Use default for republishing discovery
	weatherPub.SetDiscoveryInterval(0, weatherApp.PublishNodes)
	// Update the forecast once an hour
	weatherPub.SetPollInterval(3600, weatherApp.UpdateWeather)
	// Handle commands to delete/remove a city node
	// weatherPub.SetNodeCommandHandler(standard.CommandCreate, onAddCity)
	// weatherPub.SetNodeCommandHandler(standard.CommandDelete, onRemoveCity)
	// Handle city node configuration
	// weatherPub.SetNodeCommandHandler(standard.CommandConfig, onConfig)
	// Handle set command (n/a)
	// weatherPub.SetNodeCommandHandler(standard.CommandInput, onInput)

	weatherPub.Start(false, nil, nil)
	weatherPub.WaitForSignal()
	weatherPub.Stop()
}

// Package weather demonstrates building a simple IoTConnect publisher for weather forecasts
// This publishes the current temperature and humidity for the cities
package main

import (
	"fmt"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/hspaay/iotconnect.openweathermap/openweathermap"
	"github.com/sirupsen/logrus"
)

const loginName = "testuser"
const password = "testuser"
const clientID = "weather.test"
const mqttServerAddress = "localhost"

var cities = []string{"Amsterdam", "Vancouver"}
var ioCurrentInstance = "current"

const publisherID = "weather"
const zoneID = standard.LocalZoneID
const KelvinToC = 273.1 // Kelvin at 0 celcius

var weatherPub *publisher.PublisherState

// Discover creates the nodes and outputs
func Discover(weatherPub *publisher.PublisherState) {
	// Publisher configuration for API
	node := weatherPub.GetNodeByID(standard.PublisherNodeID)
	apikeyConfig := standard.NewConfig("apikey", standard.DataTypeString, "Weather Service API key",
		"please-register")
	apikeyConfig.Secret = true // don't publish the value
	weatherPub.DiscoverNodeConfig(node, apikeyConfig)
	// future enhancement would be to add inputs to add/remove city nodes

	// Create a node for each city with temperature outputs
	for _, city := range cities {
		weatherPub.DiscoverNode(standard.NewNode(zoneID, publisherID, city))
		weatherPub.DiscoverNodeConfig(node, standard.NewConfig("city", standard.DataTypeString, "City", city))

		// Output the current temperature and humidity
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeTemperature, ioCurrentInstance))
		weatherPub.DiscoverOutput(standard.NewOutput(node, standard.IOTypeHumidity, ioCurrentInstance))
	}
}

// Poll obtains the weather and updates the output value
// The iotconnect library will automatically publish the output discovery and values.
func Poll(weatherPub *publisher.PublisherState) {
	pubNode := weatherPub.GetNodeByID(standard.PublisherNodeID)
	apikey := pubNode.Config["apikey"].Value

	// publish the current weather for each of the city nodes
	for _, node := range weatherPub.GetAllNodes() {
		cityConfig := node.Config["city"]
		if cityConfig != nil && cityConfig.Value != "" {
			currentWeather, err := openweathermap.GetCurrentWeather(apikey, cityConfig.Value)
			if err != nil {
				weatherPub.SetErrorStatus(node, "Forecast not available")
				return
			}
			weatherPub.UpdateOutputValue(node, standard.IOTypeTemperature, ioCurrentInstance, fmt.Sprintf("%f", currentWeather.Main.Temperature+KelvinToC))
			weatherPub.UpdateOutputValue(node, standard.IOTypeHumidity, ioCurrentInstance, fmt.Sprintf("%d", currentWeather.Main.Humidity))
		}
	}
}

func main() {
	logger := logrus.New()
	messenger := messenger.NewMqttMessenger(mqttServerAddress, 0, loginName, password, clientID, logger) // use default mqtt port
	weatherPub = publisher.NewPublisher(zoneID, publisherID, messenger)

	// discover the node(s) and outputs
	Discover(weatherPub)

	// Update the forecast once an hour
	weatherPub.SetPollInterval(3600, Poll)

	weatherPub.Start(false, nil, nil)
	weatherPub.WaitForSignal()
	weatherPub.Stop()
}

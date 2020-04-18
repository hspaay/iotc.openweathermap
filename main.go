package main

import (
	"github.com/hspaay/iotconnect.golang/config"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.openweathermap/openweathermap"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	messengerConfig := &messenger.MessengerConfig{}
	weatherApp := openweathermap.NewWeatherApp()
	config.LoadAppConfig("", weatherApp.PublisherID, messengerConfig, &weatherApp)

	messenger := messenger.NewMqttMessenger(messengerConfig, logger)
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, messenger)

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

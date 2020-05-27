package main

import (
	"github.com/hspaay/iotc.golang/persist"
	"github.com/hspaay/iotc.golang/publisher"
	"github.com/hspaay/iotc.openweathermap/internal"
)

func main() {
	configFolder := persist.DefaultConfigFolder
	weatherApp := internal.NewWeatherApp()
	weatherPub, _ := publisher.NewAppPublisher("openweathermap", configFolder, &weatherApp, true)

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

package openweathermap

import (
	"testing"
	"time"

	"github.com/hspaay/iotconnect.golang/config"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/stretchr/testify/assert"
)

const zoneID = standard.LocalZoneID
const configFolder = ""

var weatherApp = WeatherApp{
	APIKey: "please register",
	Cities: []string{"Amsterdam", "Vancouver"},
}

// TestNewPublisher instance
func TestNewPublisher(t *testing.T) {

	// logger := log.New()
	// testMessenger := messenger.NewMqttMessenger(mqttServerAddress, 0, loginName, password, clientID, logger) // use default mqtt port
	// config.LoadAppConfig("", publisherID, nil, &testConfig)
	var testMessenger = messenger.NewDummyMessenger()
	weatherPub := publisher.NewPublisher(zoneID, PublisherID, testMessenger)

	weatherPub.Start(false, nil, nil)
	weatherApp.PublishNodes(weatherPub)
	// pubNode := weatherPub.GetNodeByID(standard.PublisherNodeID)
	// apikeyConfig := pubNode.Config[APIKEY_CONFIG]

	if !assert.NotNil(t, weatherApp.APIKey, "Missing apikey in configuration") {
		return
	}
	weatherPub.Stop()
}

func TestPublishWeather(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	weatherPub := publisher.NewPublisher(zoneID, PublisherID, testMessenger)

	err := config.LoadAppConfig(configFolder, PublisherID, nil, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", PublisherID, err) {
		return
	}
	weatherPub.Start(false, nil, nil)
	weatherApp.PublishNodes(weatherPub)
	weatherApp.UpdateWeather(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

func TestPublishForecast(t *testing.T) {
	var testMessenger = messenger.NewDummyMessenger()
	weatherPub := publisher.NewPublisher(zoneID, PublisherID, testMessenger)

	err := config.LoadAppConfig(configFolder, PublisherID, nil, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", PublisherID, err) {
		return
	}
	weatherPub.Start(false, nil, nil)
	weatherApp.PublishNodes(weatherPub)
	weatherApp.UpdateForecast(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

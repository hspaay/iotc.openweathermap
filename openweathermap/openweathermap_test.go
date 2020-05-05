package openweathermap

import (
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/messaging"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
	"github.com/hspaay/iotc.golang/publisher"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const zoneID = messaging.LocalZoneID
const configFolder = "./test"

var messengerConfig = &messenger.MessengerConfig{}

var weatherApp = WeatherApp{
	APIKey: "please register",
	Cities: []string{"Amsterdam", "Vancouver"},
}

// TestNewPublisher instance
func TestNewPublisher(t *testing.T) {

	// logger := log.New()
	// testMessenger := messenger.NewMqttMessenger(mqttServerAddress, 0, loginName, password, clientID, logger) // use default mqtt port
	// config.LoadAppConfig("", publisherID, nil, &testConfig)
	persist.LoadMessengerConfig(configFolder, messengerConfig)
	testMessenger := messenger.NewDummyMessenger(messengerConfig, nil)
	weatherPub := publisher.NewPublisher(PublisherID, testMessenger, configFolder)

	weatherPub.Start()
	weatherApp.PublishNodes(weatherPub)

	if !assert.NotNil(t, weatherApp.APIKey, "Missing apikey in configuration") {
		return
	}
	weatherPub.Stop()
}

func TestPublishWeather(t *testing.T) {
	persist.LoadMessengerConfig(configFolder, messengerConfig)
	testMessenger := messenger.NewDummyMessenger(messengerConfig, nil)
	weatherPub := publisher.NewPublisher(PublisherID, testMessenger, configFolder)

	err := persist.LoadAppConfig(configFolder, PublisherID, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", PublisherID, err) {
		return
	}
	weatherPub.Start()
	weatherApp.PublishNodes(weatherPub)
	weatherApp.UpdateWeather(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

func TestPublishForecast(t *testing.T) {
	persist.LoadMessengerConfig(configFolder, messengerConfig)
	testMessenger := messenger.NewDummyMessenger(messengerConfig, nil)
	weatherPub := publisher.NewPublisher(PublisherID, testMessenger, configFolder)

	err := persist.LoadAppConfig(configFolder, PublisherID, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", PublisherID, err) {
		return
	}
	weatherPub.Start()
	weatherApp.PublishNodes(weatherPub)
	weatherApp.UpdateForecast(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

func TestMain(t *testing.T) {
	logger := logrus.New()

	weatherApp := NewWeatherApp()
	persist.LoadAppConfig("", weatherApp.PublisherID, &weatherApp)
	persist.LoadMessengerConfig(configFolder, messengerConfig)

	messenger := messenger.NewMqttMessenger(messengerConfig, logger)
	weatherPub := publisher.NewPublisher(weatherApp.PublisherID, messenger, configFolder)

	// Discover the node(s) and outputs. Use default for republishing discovery
	weatherPub.SetDiscoveryInterval(0, weatherApp.PublishNodes)
	weatherPub.SetPollInterval(30, weatherApp.UpdateWeather)

	weatherPub.Start()
	time.Sleep(time.Minute * 60)
	weatherPub.Stop()
}

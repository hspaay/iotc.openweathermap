package internal

import (
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/messenger"
	"github.com/hspaay/iotc.golang/persist"
	"github.com/hspaay/iotc.golang/publisher"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const zoneID = iotc.TestZoneID
const configFolder = "../test"

var messengerConfig = &messenger.MessengerConfig{Zone: zoneID}

var weatherApp = WeatherApp{
	APIKey:      "please register",
	Cities:      []string{"Amsterdam", "Vancouver"},
	PublisherID: AppID,
}

// TestNewPublisher instance
func TestNewPublisher(t *testing.T) {

	// logger := log.New()
	// testMessenger := messenger.NewMqttMessenger(mqttServerAddress, 0, loginName, password, clientID, logger) // use default mqtt port
	// config.LoadAppConfig("", publisherID, nil, &testConfig)
	persist.LoadMessengerConfig(configFolder, messengerConfig)
	testMessenger := messenger.NewDummyMessenger(messengerConfig, nil)
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, testMessenger)
	// weatherPub.PersistNodes(configFolder, false)

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
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, testMessenger)
	// weatherPub.PersistNodes(configFolder, false)

	err := persist.LoadAppConfig(configFolder, AppID, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", AppID, err) {
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
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, testMessenger)
	// weatherPub.PersistNodes(configFolder, false)

	err := persist.LoadAppConfig(configFolder, AppID, &weatherApp)
	if !assert.NoErrorf(t, err, "Missing app configuration for publisher %s: %s", AppID, err) {
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
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, messenger)
	// weatherPub.PersistNodes(configFolder, false)

	// Discover the node(s) and outputs. Use default for republishing discovery
	weatherPub.SetDiscoveryInterval(0, weatherApp.PublishNodes)
	weatherPub.SetPollInterval(30, weatherApp.UpdateWeather)

	weatherPub.Start()
	time.Sleep(time.Minute * 60)
	weatherPub.Stop()
}

package openweathermap

import (
	"testing"
	"time"

	"github.com/hspaay/iotconnect.golang/config"
	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/sirupsen/logrus"
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

	weatherPub.Start()
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
	weatherPub.Start()
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
	weatherPub.Start()
	weatherApp.PublishNodes(weatherPub)
	weatherApp.UpdateForecast(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

func TestMain(t *testing.T) {
	logger := logrus.New()

	messengerConfig := &messenger.MessengerConfig{}
	weatherApp := NewWeatherApp()
	config.LoadAppConfig("", weatherApp.PublisherID, messengerConfig, &weatherApp)

	messenger := messenger.NewMqttMessenger(messengerConfig, logger)
	weatherPub := publisher.NewPublisher(messengerConfig.Zone, weatherApp.PublisherID, messenger)

	// Discover the node(s) and outputs. Use default for republishing discovery
	weatherPub.SetDiscoveryInterval(0, weatherApp.PublishNodes)
	weatherPub.SetPollInterval(30, weatherApp.UpdateWeather)

	weatherPub.Start()
	time.Sleep(time.Minute * 60)
	weatherPub.Stop()
}

package internal

import (
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
)

const domain = types.TestDomainID
const cacheFolder = "../test/cache"
const configFolder = "../test"

var messengerConfig = &messaging.MessengerConfig{Domain: domain}

var weatherApp = WeatherApp{
	APIKey:      "please register",
	Cities:      []string{"Amsterdam", "Vancouver"},
	PublisherID: AppID,
}

// TestNewPublisher instance
func TestNewPublisher(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, &weatherApp, "", false)
	assert.NoErrorf(t, err, "error in NewAppPublisher")
	assert.NotNil(t, weatherApp.APIKey, "Missing apikey in configuration")
	assert.NotEmptyf(t, pub.PublisherID(), "Missing publisher ID")

	pub.Start()
	weatherApp.PublishNodes(pub)
	pub.Stop()
}

func TestPublishWeather(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, &weatherApp, "", false)
	assert.NoErrorf(t, err, "error in NewAppPublisher")
	assert.NotNil(t, weatherApp.APIKey, "Missing apikey in configuration")

	pub.Start()
	weatherApp.PublishNodes(pub)
	weatherApp.UpdateWeather(pub)

	time.Sleep(time.Second * 3)
	pub.Stop()
}

func TestPublishForecast(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, &weatherApp, "", false)
	assert.NoErrorf(t, err, "error in NewAppPublisher")

	pub.Start()
	weatherApp.PublishNodes(pub)
	weatherApp.UpdateForecast(pub)

	time.Sleep(time.Second * 3)
	pub.Stop()
}

func TestMain(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, &weatherApp, "", false)
	assert.NoErrorf(t, err, "error in NewAppPublisher")

	// Discover the node(s) and outputs. Use default for republishing discovery
	// pub.SetDiscoyInterval(0, weatherApp.PublishNodes)
	pub.SetPollInterval(30, weatherApp.UpdateWeather)

	pub.Start()
	time.Sleep(time.Second * 40)
	pub.Stop()
}

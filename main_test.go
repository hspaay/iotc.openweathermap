package main

import (
	"testing"
	"time"

	"github.com/hspaay/iotconnect.golang/messenger"
	"github.com/hspaay/iotconnect.golang/publisher"
	"github.com/hspaay/iotconnect.golang/standard"
	"github.com/stretchr/testify/assert"
)

const APIKEY_CONFIG = "apikey"

// TestNewPublisher instance
func TestNewPublisher(t *testing.T) {

	// logger := log.New()
	// testMessenger := messenger.NewMqttMessenger(mqttServerAddress, 0, loginName, password, clientID, logger) // use default mqtt port
	var testMessenger = messenger.NewDummyMessenger()
	weatherPub := publisher.NewPublisher(zoneID, publisherID, testMessenger)

	weatherPub.Start(false, nil, nil)
	Discover(weatherPub)
	pubNode := weatherPub.GetNodeByID(standard.PublisherNodeID)
	apikeyConfig := pubNode.Config[APIKEY_CONFIG]

	if !assert.NotNil(t, apikeyConfig, "Missing apikey configuration in publisher") {
		return
	}

	Poll(weatherPub)

	time.Sleep(time.Second * 3)
	weatherPub.Stop()
}

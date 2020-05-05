# IoTConnect OpenWeatherMap

This demonstration project written in golang publishes weather information onto an MQTT message bus following the IoTConnect standard. It obtains current weather using openweathermap.

## Status
Basic functional source.
work in progress...

## Configuration

Two configuration files are expected:
1. ~/bin/iotconnect/config/iotc.conf which is used by with all publishers
2. ~/bin/iotconnect/config/openweathermap.conf with api key and cities

See config files in ./test as examples
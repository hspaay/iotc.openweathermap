module github.com/hspaay/iotc.openweathermap

go 1.13

require (
	github.com/hspaay/iotc.golang v0.0.0-20200418081053-7ded037c794f
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.0
)

// Temporary for testing iotc.golang
replace github.com/hspaay/iotc.golang => ../iotc.golang

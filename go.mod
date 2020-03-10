module github.com/hspaay/iotconnect.openweathermap

go 1.13

require (
	github.com/hspaay/iotconnect.golang v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
)

// Temporary for testing iotconnect.golang
replace github.com/hspaay/iotconnect.golang => ../iotconnect.golang

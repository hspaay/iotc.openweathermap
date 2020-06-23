module openweathermap

go 1.13

require (
	github.com/iotdomain/iotdomain-go v0.0.0-20200418081053-7ded037c794f
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.0
)

// Temporary for testing iotdomain-go
replace github.com/iotdomain/iotdomain-go => ../iotdomain-go

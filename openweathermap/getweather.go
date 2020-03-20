package openweathermap

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// Sign up to openweathermap.org to obtain an api key for your app"
const currentWeatherURL = "https://api.openweathermap.org/data/2.5/weather?q={city}&appid={apikey}&units=metric&lang={lang}"
const forecastURL = "https://api.openweathermap.org/data/2.5/forecast?q={city}&appid={apikey}&units=metric&lang={lang}"

// CurrentWeather API result
type CurrentWeather struct {
	Coord struct {
		Lat float32 `json:"lat"`
		Lon float32 `json:"lon"`
	} `json:"coord"`
	Main struct {
		Humidity     int     `json:"humidity"`   // Default Kelvin
		FeelsLike    float32 `json:"feels_like"` // HeatIndex, Default Kelvin
		Pressure     float32 `json:"pressure"`   // atmospheric pressure hPa
		PressureGrnd float32 `json:"grnd_level"` // atmospheric pressure hPa
		PressureSea  float32 `json:"sea_level"`  // atmospheric pressure hPa
		Temperature  float32 `json:"temp"`       // Default Kelvin
		TempMax      float32 `json:"temp_max"`   // Max in area. Default Kelvin
		TempMin      float32 `json:"temp_min"`   // Min in area. Default Kelvin
	} `json:"main"`
	Name string `json:"name"` // city name
	Rain struct {
		LastHour   float32 `json:"1h"` // rainfall in the last hour in mm
		Last3Hours float32 `json:"3h"` // rainfall in the last 3 hour in mm
	} `json:"rain"`
	Sys struct {
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Snow struct {
		LastHour   float32 `json:"1h"` // snowfall in the last hour in mm
		Last3Hours float32 `json:"3h"` // snowfall in the last 3 hour in mm
	} `json:"snow"`
	Timestamp int `json:"dt"`       // in UTC
	TimeZone  int `json:"timezone"` // time offset from UTC in seconds
	Weather   []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed   float32 `json:"speed"` // Default: m/s
		Heading float32 `json:"deg"`   // Default degrees
	} `json:"wind"`
}

// ForecastMessage containing 5 day forecast
type ForecastMessage struct {
	Count int `json:"cnt"`
	List  []struct {
		Clouds struct {
			Percent int `json:"all"`
		} `json:"clouds"`
		Date     int    `json:"dt"`
		DateText string `json:"dt_txt"`
		Main     struct {
			FeelsLike    float32 `json:"feels_like"` // HeatIndex, Default Kelvin
			Humidity     int     `json:"humidity"`   // Default Kelvin
			Pressure     float32 `json:"pressure"`   // atmospheric pressure hPa
			PressureGrnd float32 `json:"grnd_level"` // atmospheric pressure hPa
			PressureSea  float32 `json:"sea_level"`  // atmospheric pressure hPa
			Temperature  float32 `json:"temp"`       // Default Kelvin
			TempMax      float32 `json:"temp_max"`   // Max in area. Default Kelvin
			TempMin      float32 `json:"temp_min"`   // Min in area. Default Kelvin
		} `json:"main"`
		Weather []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Wind struct {
			Speed   float32 `json:"speed"` // Default: m/s
			Heading float32 `json:"deg"`   // Default degrees
		} `json:"wind"`
	}
}

// GetCurrentWeather reads the current weather from the openweathermap service
func GetCurrentWeather(apikey string, city string, lang string) (*CurrentWeather, error) {

	requestURL := strings.Replace(currentWeatherURL, "{apikey}", apikey, -1)
	requestURL = strings.Replace(requestURL, "{city}", city, -1)
	requestURL = strings.Replace(requestURL, "{lang}", lang, -1)

	var currentWeather *CurrentWeather

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("Invalid API key")
	} else if resp.StatusCode >= 400 {
		return nil, errors.New("Request failed")
	}
	currentRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	weatherStr := string(currentRaw)
	_ = weatherStr
	json.Unmarshal(currentRaw, &currentWeather)
	return currentWeather, nil
}

// Get5DayForecast reads the 5 day forecast from the openweathermap service
func Get5DayForecast(apikey string, city string, lang string) (*ForecastMessage, error) {

	requestURL := strings.Replace(forecastURL, "{apikey}", apikey, -1)
	requestURL = strings.Replace(requestURL, "{city}", city, -1)
	requestURL = strings.Replace(requestURL, "{lang}", lang, -1)

	var forecastWeather *ForecastMessage

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("Invalid API key")
	} else if resp.StatusCode >= 400 {
		return nil, errors.New("Request failed")
	}
	forecastRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	weatherStr := string(forecastRaw)
	_ = weatherStr
	json.Unmarshal(forecastRaw, &forecastWeather)
	return forecastWeather, nil
}

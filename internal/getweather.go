package internal

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// Sign up to openweathermap.org to obtain an api key for your app"
const currentWeatherURL = "https://api.openweathermap.org/data/2.5/weather?q={city}&appid={apikey}&units=metric&lang={lang}"
const threeHourlyForecastURL = "https://api.openweathermap.org/data/2.5/forecast?q={city}&appid={apikey}&units=metric&lang={lang}"
const dailyForecastURL = "https://api.openweathermap.org/data/2.5/daily?q={city}&appid={apikey}&units=metric&lang={lang}"

// CurrentWeather API result
type CurrentWeather struct {
	Coord struct {
		Lat float32 `json:"lat"`
		Lon float32 `json:"lon"`
	} `json:"coord"`
	Main struct {
		Humidity     int     `json:"humidity"`
		FeelsLike    float32 `json:"feels_like"` // HeatIndex
		Pressure     float32 `json:"pressure"`   // atmospheric pressure hPa
		PressureGrnd float32 `json:"grnd_level"` // atmospheric pressure hPa
		PressureSea  float32 `json:"sea_level"`  // atmospheric pressure hPa
		Temperature  float32 `json:"temp"`       //
		TempMax      float32 `json:"temp_max"`   // Max in area
		TempMin      float32 `json:"temp_min"`   // Min in area
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
			FeelsLike    float32 `json:"feels_like"` // HeatIndex
			Humidity     int     `json:"humidity"`   //
			Pressure     float32 `json:"pressure"`   // atmospheric pressure hPa
			PressureGrnd float32 `json:"grnd_level"` // atmospheric pressure hPa
			PressureSea  float32 `json:"sea_level"`  // atmospheric pressure hPa
			Temperature  float32 `json:"temp"`       //
			TempMax      float32 `json:"temp_max"`   // Max in area
			TempMin      float32 `json:"temp_min"`   // Min in area
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
	} `json:"list"`
}

// DailyForecastMessage containing 16 days daily forecast
type DailyForecastMessage struct {
	City struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Coord struct {
			Longitude float32 `json:"long"`
			Latitude  float32 `json:"lat"`
		} `json:"coord"`
		Country  string `json:"country"`
		Timezone int    `json:"timezone"` // shift from UTC in seconds
	} `json:"city"`
	List []struct {
		Clouds   int     `json:"clouds"` // Cloudiness %
		Date     int     `json:"dt"`
		Humidity int     `json:"humidity"` //
		Pressure float32 `json:"pressure"` // sealevel atmospheric pressure hPa
		Rain     float32 `json:"rain"`     // rain in mm
		Snow     float32 `json:"snow"`     // snow in mm
		Sunrise  int     `json:"sunrise"`
		Sunset   int     `json:"sunset"`
		Temp     struct {
			Day     float32 `json:"day"`
			Max     float32 `json:"max"` // Daily max
			Min     float32 `json:"min"` // Daily min
			Night   float32 `json:"night"`
			Evening float32 `json:"eve"`
			Morning float32 `json:"morn"`
		} `json:"temp"`
		Weather []struct {
			ID          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		WindSpeed   float32 `json:"speed"`
		WindHeading int     `json:"deg"` // wind direction in degrees
	} `json:"list"`
}

// Call the get weather API
func getWeather(baseURL string, apikey string, city string, lang string) ([]byte, error) {
	requestURL := strings.Replace(baseURL, "{apikey}", apikey, -1)
	requestURL = strings.Replace(requestURL, "{city}", city, -1)
	requestURL = strings.Replace(requestURL, "{lang}", lang, -1)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("Invalid API key")
	} else if resp.StatusCode >= 400 {
		return nil, errors.New("Request failed")
	}
	forecastRaw, err := ioutil.ReadAll(resp.Body)
	return forecastRaw, err
}

// GetCurrentWeather reads the current weather from the openweathermap service
func GetCurrentWeather(apikey string, city string, lang string) (*CurrentWeather, error) {

	rawWeather, err := getWeather(currentWeatherURL, apikey, city, lang)
	if err != nil {
		return nil, err
	}

	var currentWeather *CurrentWeather
	json.Unmarshal(rawWeather, &currentWeather)
	return currentWeather, nil
}

// Get5DayForecast reads the 5 day forecast from the openweathermap service
func Get5DayForecast(apikey string, city string, lang string) (*ForecastMessage, error) {

	rawWeather, err := getWeather(currentWeatherURL, apikey, city, lang)
	if err != nil {
		return nil, err
	}

	var forecastWeather *ForecastMessage
	json.Unmarshal(rawWeather, &forecastWeather)
	return forecastWeather, nil
}

// GetDailyForecast reads the 16 day forecast from the openweathermap service
func GetDailyForecast(apikey string, city string, lang string) (*DailyForecastMessage, error) {

	rawWeather, err := getWeather(currentWeatherURL, apikey, city, lang)
	if err != nil {
		return nil, err
	}
	var dailyForecast *DailyForecastMessage

	json.Unmarshal(rawWeather, &dailyForecast)

	return dailyForecast, nil
}

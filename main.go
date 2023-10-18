/*

Ändra namn till weather? Så det matchar exemplet ovan

*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
)

type GeoCode struct {
	Lat          string
	Lon          string
	Display_name string
}

type Forecast struct {
	Timestamp      string
	AirTemperature float32
	WindSpeed      float32
	Precipitation  float32
}

func fetch(userAgent string, url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func fetchGeoCodeList(userAgent string, geoCodeApiUrl string, city string) ([]GeoCode, error) {
	body, err := fetch(userAgent, fmt.Sprintf("%s?q=%s", geoCodeApiUrl, city))
	if err != nil {
		return nil, err
	}

	var resultList []GeoCode
	jsonError := json.Unmarshal(body, &resultList)
	if jsonError != nil {
		return nil, jsonError
	}
	return resultList, nil
}

func fetchForecast(userAgent string, forecastApiUrl string, lat string, lon string) (Forecast, error) {

	type ApiResponse struct {
		Properties struct {
			Timeseries []struct {
				Time string
				Data struct {
					Instant struct {
						Details struct {
							Air_temperature float32
							Wind_speed      float32
						}
					}
					Next_1_hours struct {
						Details struct {
							Precipitation_amount float32
						}
					}
				}
			}
		}
	}

	body, err := fetch(userAgent, fmt.Sprintf("%s?lat=%s&lon=%s", forecastApiUrl, lat, lon))
	if err != nil {
		return Forecast{}, err
	}

	var apiResponse ApiResponse
	jsonError := json.Unmarshal(body, &apiResponse)

	if jsonError != nil || len(apiResponse.Properties.Timeseries) == 0 {
		return Forecast{}, errors.New("Unexpected API format")
	}

	return Forecast{
		Timestamp:      apiResponse.Properties.Timeseries[0].Time,
		WindSpeed:      apiResponse.Properties.Timeseries[0].Data.Instant.Details.Wind_speed,
		Precipitation:  apiResponse.Properties.Timeseries[0].Data.Next_1_hours.Details.Precipitation_amount,
		AirTemperature: apiResponse.Properties.Timeseries[0].Data.Instant.Details.Air_temperature,
	}, nil
}

func main() {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("Error reading .env file")
	}
	userAgent := viper.GetString("USERAGENT")
	geoCodeApiUrl := viper.GetString("GEOCODEAPIURL")
	forecastApiUrl := viper.GetString("FORECASTAPIURL")

	city, err := getLocationParameter()
	if err != nil {
		fmt.Println(`Missing parameter. Try running the following: ./weather "Stockholm, Stockholms kommun, Stockholm County, 111 29, Sweden"`)
		return
	}

	geoCodeList, err := fetchGeoCodeList(userAgent, geoCodeApiUrl, city)
	if err != nil {
		log.Fatalln("An error occured while fetching geo codes")
	}

	if len(geoCodeList) == 0 {
		fmt.Println("Found no matches, try again with a different parameter.")
		return
	} else if len(geoCodeList) > 1 {
		fmt.Println("Found multiple matches for your search. Try again with one of the following:")
		for _, val := range geoCodeList {
			fmt.Println(val.Display_name)
		}
		return
	}

	geoCode := geoCodeList[0]
	forecast, err := fetchForecast(userAgent, forecastApiUrl, geoCode.Lat, geoCode.Lon)

	if err != nil {
		log.Fatalln("An error occured while fetching forecast")
	}

	fmt.Println(fmt.Sprintf("Expected forecast for %s", city))
	fmt.Println(fmt.Sprintf("At %s", forecast.Timestamp))
	fmt.Println(fmt.Sprintf("Temp: %s°C", floatToString(forecast.AirTemperature)))
	fmt.Println(fmt.Sprintf("Windspeed: %sm/s", floatToString(forecast.WindSpeed)))
	fmt.Println(fmt.Sprintf("Precipitation within the next hour: %smm", floatToString(forecast.Precipitation)))
}

func floatToString(data float32) string {
	return fmt.Sprintf("%v", data)
}

func getLocationParameter() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("Missing parameter")
	}
	return os.Args[1], nil
}

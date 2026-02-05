package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ---------- Open-Meteo: Geocoding ----------
type GeoResponse struct {
	Results []GeoResult `json:"results"`
}

type GeoResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ---------- Open-Meteo: Weather ----------
type OpenMeteoResponse struct {
	Current Current `json:"current"`
}

type Current struct {
	Temperature2m       float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	RelativeHumidity2m  int     `json:"relative_humidity_2m"`
	PrecipProbability   int     `json:"precipitation_probability"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	WeatherCode         int     `json:"weather_code"`
}

// ---------- Air Quality ----------
type AirQualityResponse struct {
	Current AirQualityCurrent `json:"current"`
}

type AirQualityCurrent struct {
	PM10  float64 `json:"pm10"`
	PM25  float64 `json:"pm2_5"`
	AQIUS int     `json:"us_aqi"`
}

// RunNow: city -> lat/lon -> weather -> print
func RunNow(city string) error {
	client := &http.Client{Timeout: 8 * time.Second}

	loc, err := geocode(client, city)
	if err != nil {
		return err
	}

	w, err := fetchCurrentWeather(client, loc.Latitude, loc.Longitude)
	if err != nil {
		return err
	}

	aq, err := fetchAirQuality(client, loc.Latitude, loc.Longitude)
	if err != nil {
		return err
	}

	now := time.Now().In(time.FixedZone("KST", 9*60*60))

	fmt.Printf("%s, %s | %s (KST)\n", loc.Name, loc.Country, now.Format("2006-01-02 15:04"))
	fmt.Printf("%s  %.1f°C (체감 %.1f°C)\n",
		iconForCode(w.WeatherCode),
		w.Temperature2m,
		w.ApparentTemperature,
	)
	fmt.Printf("강수확률 %d%% | 습도 %d%% | 바람 %.1f m/s\n",
		w.PrecipProbability,
		w.RelativeHumidity2m,
		w.WindSpeed10m,
	)
	fmt.Printf("미세먼지 PM2.5 %.1f ㎍/m³ | PM10 %.1f ㎍/m³ | AQI %d (%s)\n",
		aq.PM25,
		aq.PM10,
		aq.AQIUS,
		aqiStatus(aq.AQIUS),
	)

	return nil
}

func geocode(client *http.Client, city string) (GeoResult, error) {
	q := url.QueryEscape(city)
	u := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ko&format=json", q)

	resp, err := client.Get(u)
	if err != nil {
		return GeoResult{}, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GeoResult{}, fmt.Errorf("geocoding bad status: %s", resp.Status)
	}

	var gr GeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return GeoResult{}, fmt.Errorf("geocoding decode failed: %w", err)
	}

	if len(gr.Results) == 0 {
		return GeoResult{}, fmt.Errorf("no results for city: %q", city)
	}

	return gr.Results[0], nil
}

func fetchCurrentWeather(client *http.Client, lat, lon float64) (Current, error) {
	u := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&timezone=Asia%%2FSeoul&current=temperature_2m,apparent_temperature,relative_humidity_2m,precipitation_probability,wind_speed_10m,weather_code",
		lat, lon,
	)

	resp, err := client.Get(u)
	if err != nil {
		return Current{}, fmt.Errorf("weather request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Current{}, fmt.Errorf("weather bad status: %s", resp.Status)
	}

	var data OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Current{}, fmt.Errorf("weather decode failed: %w", err)
	}

	return data.Current, nil
}

func fetchAirQuality(client *http.Client, lat, lon float64) (AirQualityCurrent, error) {
	u := fmt.Sprintf(
		"https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%f&longitude=%f&timezone=Asia%%2FSeoul&current=pm10,pm2_5,us_aqi",
		lat, lon,
	)

	resp, err := client.Get(u)
	if err != nil {
		return AirQualityCurrent{}, fmt.Errorf("air quality request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AirQualityCurrent{}, fmt.Errorf("air quality bad status: %s", resp.Status)
	}

	var data AirQualityResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return AirQualityCurrent{}, fmt.Errorf("air quality decode failed: %w", err)
	}

	return data.Current, nil
}

// --- helpers ---
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func iconForCode(code int) string {
	switch code {
	case 0:
		return "☀️ 맑음"
	case 1, 2, 3:
		return "☁️ 흐림"
	case 45, 48:
		return "🌫️ 안개"
	case 51, 53, 55:
		return "🌦️ 이슬비"
	case 61, 63, 65:
		return "🌧️ 비"
	case 71, 73, 75:
		return "🌨️ 눈"
	case 95:
		return "⛈️ 뇌우"
	default:
		return "🌡️ 알 수 없음"
	}
}

func aqiStatus(aqi int) string {
	switch {
	case aqi <= 50:
		return "좋음 😊"
	case aqi <= 100:
		return "보통 🙂"
	case aqi <= 150:
		return "나쁨 😷"
	case aqi <= 200:
		return "매우 나쁨 🤢"
	default:
		return "위험 ☠️"
	}
}
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"NexusAi/common/mcp/base"
)

// AmapWeatherResponse 高德天气 API 响应
type AmapWeatherResponse struct {
	Status   string `json:"status"`
	Count    string `json:"count"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Lives    []struct {
		Province      string `json:"province"`
		City          string `json:"city"`
		Weather       string `json:"weather"`
		Temperature   string `json:"temperature"`
		WindDirection string `json:"winddirection"`
		WindPower     string `json:"windpower"`
		Humidity      string `json:"humidity"`
		ReportTime    string `json:"reporttime"`
	} `json:"lives"`
}

// AmapGeocodeResponse 高德城市搜索响应
type AmapGeocodeResponse struct {
	Status   string `json:"status"`
	Count    string `json:"count"`
	Geocodes []struct {
		Adcode string `json:"adcode"`
		Name   string `json:"name"`
		City   string `json:"city"`
	} `json:"geocodes"`
}

// WeatherApiClient 天气 API 客户端
type WeatherApiClient struct {
	httpClient *http.Client
	apiKey     string
}

// NewWeatherApiClient 创建天气 API 客户端
func NewWeatherApiClient() *WeatherApiClient {
	apiKey := os.Getenv("AMAP_API_KEY")
	return &WeatherApiClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey: apiKey,
	}
}

// getAdcode 获取城市的 adcode（行政区划代码）
func (c *WeatherApiClient) getAdcode(ctx context.Context, city string) (adcode, cityName string, err error) {
	if c.apiKey == "" {
		return "", "", fmt.Errorf("AMAP_API_KEY not configured")
	}

	apiUrl := fmt.Sprintf("https://restapi.amap.com/v3/geocode/geo?address=%s&key=%s",
		url.QueryEscape(city), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create geocode request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to call geocode API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read geocode API response: %w", err)
	}

	var geocodeResp AmapGeocodeResponse
	if err := json.Unmarshal(body, &geocodeResp); err != nil {
		return "", "", fmt.Errorf("failed to parse geocode API response: %w", err)
	}

	if geocodeResp.Status != "1" || len(geocodeResp.Geocodes) == 0 {
		return "", "", fmt.Errorf("city not found: %s (status: %s, count: %s)", city, geocodeResp.Status, geocodeResp.Count)
	}

	name := geocodeResp.Geocodes[0].Name
	if name == "" {
		name = geocodeResp.Geocodes[0].City
	}

	return geocodeResp.Geocodes[0].Adcode, name, nil
}

// GetWeather 获取天气信息
func (c *WeatherApiClient) GetWeather(ctx context.Context, location string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("AMAP_API_KEY not configured")
	}

	// 1. 获取城市的 adcode
	adcode, cityName, err := c.getAdcode(ctx, location)
	if err != nil {
		return "", err
	}

	// 2. 获取实时天气
	apiUrl := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?city=%s&key=%s&extensions=base",
		adcode, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create weather request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call weather API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read weather API response: %w", err)
	}

	var weatherResp AmapWeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return "", fmt.Errorf("failed to parse weather API response: %w", err)
	}

	if weatherResp.Status != "1" || len(weatherResp.Lives) == 0 {
		return "", fmt.Errorf("weather API error: status %s, count %s", weatherResp.Status, weatherResp.Count)
	}

	live := weatherResp.Lives[0]
	temp, _ := strconv.ParseFloat(live.Temperature, 64)
	humidity, _ := strconv.Atoi(live.Humidity)

	return fmt.Sprintf(
		"城市: %s\n温度: %.1f°C\n天气: %s\n湿度: %d%%\n风向: %s\n风力: %s级",
		cityName,
		temp,
		live.Weather,
		humidity,
		live.WindDirection,
		live.WindPower,
	), nil
}

//MCP 服务创建

// GetWeatherServiceConfig 获取天气服务配置
func GetWeatherServiceConfig() base.ServiceConfig {
	weatherClient := NewWeatherApiClient()

	return base.ServiceConfig{
		Name:    "weather",
		Version: "1.0.0",
		Tools: []base.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "获取指定位置的当前天气信息",
				Parameters: []base.ToolParameter{
					{
						Name:        "location",
						Description: "要查询天气的地理位置，例如城市名称（北京、上海、广州等）",
						Required:    true,
					},
				},
				Handler: func(ctx context.Context, args map[string]any) (string, error) {
					location, ok := args["location"].(string)
					if !ok || location == "" {
						return "", fmt.Errorf("invalid argument: location is required and must be a string")
					}
					return weatherClient.GetWeather(ctx, location)
				},
			},
		},
	}
}

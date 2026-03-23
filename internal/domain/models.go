package domain

import "time"

type Metric struct {
	Value float64
	Index int
}

type Device struct {
	Name       string `json:"name"`
	DeviceType string `json:"deviceType"`
	DeviceID   int    `json:"deviceId"`
}

type Reading struct {
	Timestamp time.Time
	Score     float64
	Temp      Metric
	Humidity  Metric
	CO2       Metric
	VOC       Metric
	PM25      Metric
}

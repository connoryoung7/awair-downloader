package awair

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/connoryoung/awair-downloader/internal/domain"
)

const baseURL = "https://developer-apis.awair.is/v1"

type apiComponent struct {
	Comp  string  `json:"comp"`
	Value float64 `json:"value"`
}

type apiReading struct {
	Timestamp time.Time      `json:"timestamp"`
	Sensors   []apiComponent `json:"sensors"`
	Indices   []apiComponent `json:"indices"`
	Score     float64        `json:"score"`
}

func (a *apiReading) toDomain() domain.Reading {
	r := domain.Reading{Timestamp: a.Timestamp, Score: a.Score}
	for _, s := range a.Sensors {
		switch s.Comp {
		case "temp":
			r.Temp.Value = s.Value
		case "humid":
			r.Humidity.Value = s.Value
		case "co2":
			r.CO2.Value = s.Value
		case "voc":
			r.VOC.Value = s.Value
		case "pm25":
			r.PM25.Value = s.Value
		}
	}
	for _, idx := range a.Indices {
		switch idx.Comp {
		case "temp":
			r.Temp.Index = int(idx.Value)
		case "humid":
			r.Humidity.Index = int(idx.Value)
		case "co2":
			r.CO2.Index = int(idx.Value)
		case "voc":
			r.VOC.Index = int(idx.Value)
		case "pm25":
			r.PM25.Index = int(idx.Value)
		}
	}
	return r
}

type Client struct {
	token string
	http  http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		http:  http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Devices() ([]domain.Device, error) {
	var result struct {
		Devices []domain.Device `json:"devices"`
	}
	if err := c.get("/users/self/devices", &result); err != nil {
		return nil, err
	}
	return result.Devices, nil
}

func (c *Client) Latest(deviceType string, deviceID int) (*domain.Reading, error) {
	path := fmt.Sprintf("/users/self/devices/%s/%d/air-data/latest", deviceType, deviceID)
	var result struct {
		Data []apiReading `json:"data"`
	}
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no data returned")
	}
	r := result.Data[0].toDomain()
	return &r, nil
}

// RawData fetches all raw readings between from and to, handling pagination.
func (c *Client) RawData(deviceType string, deviceID int, from, to time.Time) ([]domain.Reading, error) {
	const limit = 360
	var all []domain.Reading
	cursor := from

	for {
		path := fmt.Sprintf("/users/self/devices/%s/%d/air-data/raw?%s",
			deviceType, deviceID,
			url.Values{
				"from":  []string{cursor.UTC().Format(time.RFC3339)},
				"to":    []string{to.UTC().Format(time.RFC3339)},
				"limit": []string{fmt.Sprint(limit)},
			}.Encode(),
		)

		var result struct {
			Data []apiReading `json:"data"`
		}
		if err := c.get(path, &result); err != nil {
			return nil, err
		}

		for _, a := range result.Data {
			all = append(all, a.toDomain())
		}

		if len(result.Data) < limit {
			break
		}
		cursor = result.Data[len(result.Data)-1].Timestamp.Add(time.Second)
		if !cursor.Before(to) {
			break
		}
	}

	return all, nil
}

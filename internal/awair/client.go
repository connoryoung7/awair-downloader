package awair

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://developer-apis.awair.is/v1"

type Device struct {
	Name       string `json:"name"`
	DeviceType string `json:"deviceType"`
	DeviceID   int    `json:"deviceId"`
}

type component struct {
	Comp  string  `json:"comp"`
	Value float64 `json:"value"`
}

type Reading struct {
	Timestamp time.Time   `json:"timestamp"`
	Score     float64     `json:"score"`
	Sensors   []component `json:"sensors"`
	Indices   []component `json:"indices"`
}

// Sensor returns the raw sensor value for the given component name (e.g. "temp").
func (r *Reading) Sensor(comp string) (float64, bool) {
	for _, s := range r.Sensors {
		if s.Comp == comp {
			return s.Value, true
		}
	}
	return 0, false
}

// Index returns the index value for the given component name.
func (r *Reading) Index(comp string) (float64, bool) {
	for _, s := range r.Indices {
		if s.Comp == comp {
			return s.Value, true
		}
	}
	return 0, false
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

func (c *Client) Devices() ([]Device, error) {
	var result struct {
		Devices []Device `json:"devices"`
	}
	if err := c.get("/users/self/devices", &result); err != nil {
		return nil, err
	}
	return result.Devices, nil
}

func (c *Client) Latest(deviceType string, deviceID int) (*Reading, error) {
	path := fmt.Sprintf("/users/self/devices/%s/%d/air-data/latest", deviceType, deviceID)
	var result struct {
		Data []Reading `json:"data"`
	}
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no data returned")
	}
	return &result.Data[0], nil
}

// RawData fetches all raw readings between from and to, handling pagination.
func (c *Client) RawData(deviceType string, deviceID int, from, to time.Time) ([]Reading, error) {
	const limit = 360
	var all []Reading
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
			Data []Reading `json:"data"`
		}
		if err := c.get(path, &result); err != nil {
			return nil, err
		}

		all = append(all, result.Data...)

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

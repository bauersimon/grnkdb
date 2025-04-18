package steam

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client is a client for the Steam web API.
type Client struct {
	baseUrl string
}

// NewClient returns a new instance.
func NewClient() *Client {
	return &Client{
		baseUrl: "https://store.steampowered.com/api",
	}
}

var steamNameCache = map[string]string{}

// GameName gets the name of a game via its AppID.
func (c *Client) GameName(appID string) (game string, err error) {
	if game = steamNameCache[appID]; game != "" {
		return game, nil
	}
	defer func() {
		if err == nil {
			steamNameCache[appID] = game
		}
	}()

	url, err := url.JoinPath(c.baseUrl, "appdetails?appids="+appID)
	if err != nil {
		return "", errors.WithStack(err)
	}
	resp, err := (&http.Client{}).Get(url)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var apiResponse map[string]struct {
		Success bool `json:"success"`
		Data    struct {
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", errors.WithStack(err)
	}

	if appData, exists := apiResponse[appID]; exists && appData.Success {
		return appData.Data.Name, nil
	}

	return "", errors.Errorf("unknown game ID %q", appID)
}

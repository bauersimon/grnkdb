package steam

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

// Client is a client for the Steam web API.
type Client struct {
	baseUrl string
}

// NewClient returns a new instance.
func NewClient() *Client {
	return &Client{
		baseUrl: "https://store.steampowered.com/api/",
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

	url, err := url.JoinPath(c.baseUrl, "appdetails")
	if err != nil {
		return "", errors.WithStack(err)
	}
	url += "?appids=" + appID

	var resp *http.Response
	var body []byte
	if err := retry.Do(func() (err error) {
		resp, err = (&http.Client{}).Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode == 429 {
			return errors.Errorf("rate limit reached: %q", string(body))
		} else if resp.StatusCode != 200 {
			return errors.Errorf("invalid API reponse (%d): %q", resp.StatusCode, string(body))
		}

		return nil
	},
		retry.Attempts(5),
		retry.Delay(time.Second*5),
		retry.RetryIf(func(err error) bool {
			return err != nil && strings.Contains(err.Error(), "rate limit")
		}),
	); err != nil {
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

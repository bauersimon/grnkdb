package steam

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameName(t *testing.T) {
	type testCase struct {
		Name string

		Server func(t *testing.T) *httptest.Server
		AppID  string

		Expected string
		Error    string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			server := tc.Server(t)
			t.Cleanup(func() {
				server.CloseClientConnections()
				server.Client()
				steamNameCache = map[string]string{}
			})

			client := NewClient()
			client.baseUrl = server.URL

			actual, err := client.GameName(tc.AppID)
			if tc.Error != "" {
				assert.ErrorContainsf(t, err, tc.Error, "game=%q", actual)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}

	validate(t, &testCase{
		Name: "Simple",

		Server: func(t *testing.T) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Truef(t, strings.HasSuffix(r.URL.Path, "appdetails"), "expected suffix \"appdetails\" on %q", r.URL.Path)
				assert.True(t, r.URL.Query().Has("appids"), "expected \"appids\" query")
				assert.Equal(t, "1234", r.URL.Query().Get("appids"), "expected \"appids=1234\" query")

				_, _ = fmt.Fprintln(w, `{"1234":{"success":true,"data":{"name":"foo"}}}`)
			}))
		},
		AppID: "1234",

		Expected: "foo",
	})

	validate(t, &testCase{
		Name: "Unsuccessful",

		Server: func(t *testing.T) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Truef(t, strings.HasSuffix(r.URL.Path, "appdetails"), "expected suffix \"appdetails\" on %q", r.URL.Path)
				assert.True(t, r.URL.Query().Has("appids"), "expected \"appids\" query")
				assert.Equal(t, "1234", r.URL.Query().Get("appids"), "expected \"appids=1234\" query")

				_, _ = fmt.Fprintln(w, `{"1234":{"success":false}}`)
			}))
		},
		AppID: "1234",

		Error: "unknown game ID",
	})

	{
		tryCount := 0
		validate(t, &testCase{
			Name: "Retry",

			Server: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Truef(t, strings.HasSuffix(r.URL.Path, "appdetails"), "expected suffix \"appdetails\" on %q", r.URL.Path)
					assert.True(t, r.URL.Query().Has("appids"), "expected \"appids\" query")
					assert.Equal(t, "1234", r.URL.Query().Get("appids"), "expected \"appids=1234\" query")

					tryCount++
					switch tryCount {
					case 1:
						w.WriteHeader(429)
						_, _ = fmt.Fprintln(w, "too many requests")
					case 2:
						_, _ = fmt.Fprintln(w, `{"1234":{"success":false}}`)
					}
				}))
			},
			AppID: "1234",

			Error: "unknown game ID",
		})
	}
}

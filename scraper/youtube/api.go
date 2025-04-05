package youtube

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// initializeService sets up the YouTube service with the given credentials.
func initializeService(ctx context.Context, apiKey string) (service *youtube.Service, err error) {
	service, err = youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return service, nil
}

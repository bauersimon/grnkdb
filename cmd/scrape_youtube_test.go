package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	mockScraper "github.com/bauersimon/grnkdb/mocks/github.com/bauersimon/grnkdb/scraper"
	"github.com/bauersimon/grnkdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrapeYoutube(t *testing.T) {
	type testCase struct {
		Name string

		Setup      func(t *testing.T, scraper *mockScraper.MockInterface)
		ChannelIDs []string

		ExpectedFiles map[string]int // filename -> video count
		Error         string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mockScraper := mockScraper.NewMockInterface(t)

			if tc.Setup != nil {
				tc.Setup(t, mockScraper)
			}

			err := scrapeYoutube(mockScraper, tmpDir, tc.ChannelIDs)

			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)

				// Verify expected files were created
				for filename, expectedCount := range tc.ExpectedFiles {
					filePath := filepath.Join(tmpDir, filename)
					require.FileExists(t, filePath)

					// Read and verify CSV content
					file, err := os.Open(filePath)
					require.NoError(t, err)
					defer func() { require.NoError(t, file.Close()) }()

					videos, err := model.VideoCSVRead(file)
					require.NoError(t, err)
					assert.Len(t, videos, expectedCount, "file %s should have %d videos", filename, expectedCount)
				}
			}
		})
	}

	validate(t, &testCase{
		Name:       "Single channel successful scrape",
		ChannelIDs: []string{"UCTEST123"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			videos := []*model.Video{
				{
					VideoID:     "video1",
					Title:       "Test Video 1",
					Description: "Test description",
					Link:        "https://www.youtube.com/watch?v=video1",
					PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST123",
					Source:      model.SourceYouTube,
				},
				{
					VideoID:     "video2",
					Title:       "Test Video 2",
					Description: "Another test description",
					Link:        "https://www.youtube.com/watch?v=video2",
					PublishedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST123",
					Source:      model.SourceYouTube,
				},
			}
			scraper.EXPECT().Videos("UCTEST123").Return(videos, nil)
		},
		ExpectedFiles: map[string]int{
			"UCTEST123.csv": 2,
		},
	})

	validate(t, &testCase{
		Name:       "Multiple channels successful scrape",
		ChannelIDs: []string{"UCTEST123", "UCTEST456"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			videos1 := []*model.Video{
				{
					VideoID:     "video1",
					Title:       "Test Video 1",
					Description: "Test description",
					Link:        "https://www.youtube.com/watch?v=video1",
					PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST123",
					Source:      model.SourceYouTube,
				},
			}
			videos2 := []*model.Video{
				{
					VideoID:     "video2",
					Title:       "Test Video 2",
					Description: "Another test description",
					Link:        "https://www.youtube.com/watch?v=video2",
					PublishedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST456",
					Source:      model.SourceYouTube,
				},
				{
					VideoID:     "video3",
					Title:       "Test Video 3",
					Description: "Yet another test description",
					Link:        "https://www.youtube.com/watch?v=video3",
					PublishedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST456",
					Source:      model.SourceYouTube,
				},
			}
			scraper.EXPECT().Videos("UCTEST123").Return(videos1, nil)
			scraper.EXPECT().Videos("UCTEST456").Return(videos2, nil)
		},
		ExpectedFiles: map[string]int{
			"UCTEST123.csv": 1,
			"UCTEST456.csv": 2,
		},
	})

	validate(t, &testCase{
		Name:       "Empty video list",
		ChannelIDs: []string{"UCTEST123"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			scraper.EXPECT().Videos("UCTEST123").Return([]*model.Video{}, nil)
		},
		ExpectedFiles: map[string]int{
			"UCTEST123.csv": 0,
		},
	})

	validate(t, &testCase{
		Name:       "Scraper returns error",
		ChannelIDs: []string{"UCTEST123"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			scraper.EXPECT().Videos("UCTEST123").Return(nil, assert.AnError)
		},
		Error: "encountered errors",
	})

	validate(t, &testCase{
		Name:       "First channel succeeds, second fails",
		ChannelIDs: []string{"UCTEST123", "UCTEST456"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			videos := []*model.Video{
				{
					VideoID:     "video1",
					Title:       "Test Video 1",
					Description: "Test description",
					Link:        "https://www.youtube.com/watch?v=video1",
					PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					ChannelID:   "UCTEST123",
					Source:      model.SourceYouTube,
				},
			}
			scraper.EXPECT().Videos("UCTEST123").Return(videos, nil)
			scraper.EXPECT().Videos("UCTEST456").Return(nil, assert.AnError)
		},
		// Function should continue processing and succeed overall
		// Only first channel's CSV should be created
		ExpectedFiles: map[string]int{
			"UCTEST123.csv": 1,
		},
		Error: "encountered errors",
	})

	validate(t, &testCase{
		Name:       "All channels fail",
		ChannelIDs: []string{"UCTEST123", "UCTEST456"},
		Setup: func(t *testing.T, scraper *mockScraper.MockInterface) {
			scraper.EXPECT().Videos("UCTEST123").Return(nil, assert.AnError)
			scraper.EXPECT().Videos("UCTEST456").Return(nil, assert.AnError)
		},
		Error: "encountered errors",
	})
}

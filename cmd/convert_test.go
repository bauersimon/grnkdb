package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	mockConverter "github.com/bauersimon/grnkdb/mocks/github.com/bauersimon/grnkdb/converter"
	"github.com/bauersimon/grnkdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConvertCSVToGames(t *testing.T) {
	type testCase struct {
		Name string

		Setup        func(t *testing.T, converter *mockConverter.MockInterface)
		PrepareFiles func(t *testing.T, dir string)

		ExpectedGames []*model.Game
		Error         string
	}

	validate := func(t *testing.T, tc *testCase) {
		t.Run(tc.Name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.json")
			mockConverter := mockConverter.NewMockInterface(t)

			if tc.PrepareFiles != nil {
				tc.PrepareFiles(t, tmpDir)
			}

			if tc.Setup != nil {
				tc.Setup(t, mockConverter)
			}

			cmd := &ConvertCommand{logger: zaptest.NewLogger(t)}
			err := cmd.convertCSVToGames(mockConverter, tmpDir, outputPath)

			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				require.NoError(t, err)

				// If ExpectedGames is nil, no output file should be created
				if tc.ExpectedGames == nil {
					assert.NoFileExists(t, outputPath)
				} else {
					// Verify output file was created and contains expected games
					require.FileExists(t, outputPath)

					file, err := os.Open(outputPath)
					require.NoError(t, err)
					defer func() { require.NoError(t, file.Close()) }()

					var actualGames []*model.Game
					err = json.NewDecoder(file).Decode(&actualGames)
					require.NoError(t, err)

					assert.Equal(t, tc.ExpectedGames, actualGames)
				}
			}
		})
	}

	{
		expectedGames := []*model.Game{
			{
				Name: "Test Game",
				Content: []*model.Content{
					{
						Link:   "https://www.youtube.com/watch?v=video1",
						Start:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Source: model.SourceYouTube,
					},
				},
			},
		}
		validate(t, &testCase{
			Name: "Single CSV file conversion",
			PrepareFiles: func(t *testing.T, dir string) {
				csvContent := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source
https://www.youtube.com/watch?v=video1,2023-01-01T12:00:00Z,Test Video 1,Test description,UCTEST123,video1,youtube
https://www.youtube.com/watch?v=video2,2023-01-02T12:00:00Z,Test Video 2,Another test description,UCTEST123,video2,youtube`

				err := os.WriteFile(filepath.Join(dir, "UCTEST123.csv"), []byte(csvContent), 0644)
				require.NoError(t, err)
			},
			Setup: func(t *testing.T, converter *mockConverter.MockInterface) {
				expectedVideos := []*model.Video{
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

				converter.EXPECT().Convert(expectedVideos).Return(expectedGames, nil)
			},
			ExpectedGames: expectedGames,
		})
	}

	{
		expectedGames := []*model.Game{
			{
				Name: "Combined Game",
				Content: []*model.Content{
					{
						Link:   "https://www.youtube.com/watch?v=video1",
						Start:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Source: model.SourceYouTube,
					},
				},
			},
		}
		validate(t, &testCase{
			Name: "Multiple CSV files conversion",
			PrepareFiles: func(t *testing.T, dir string) {
				csvContent1 := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source
https://www.youtube.com/watch?v=video1,2023-01-01T12:00:00Z,Test Video 1,Test description,UCTEST123,video1,youtube`

				csvContent2 := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source
https://www.youtube.com/watch?v=video2,2023-01-02T12:00:00Z,Test Video 2,Another test description,UCTEST456,video2,youtube`

				err := os.WriteFile(filepath.Join(dir, "UCTEST123.csv"), []byte(csvContent1), 0644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "UCTEST456.csv"), []byte(csvContent2), 0644)
				require.NoError(t, err)
			},
			Setup: func(t *testing.T, converter *mockConverter.MockInterface) {
				// Since files are read in glob order, we can't predict the exact order.
				converter.EXPECT().Convert(mock.AnythingOfType("[]*model.Video")).Return(expectedGames, nil).Run(func(videos []*model.Video) {
					assert.Len(t, videos, 2, "should have 2 videos total from both CSV files")
				})
			},
			ExpectedGames: expectedGames,
		})
	}

	validate(t, &testCase{
		Name: "No CSV files in directory",
		// Don't prepare any files - function should return nil early
		// No output file should be created since there are no CSV files
		ExpectedGames: nil,
	})

	validate(t, &testCase{
		Name: "Empty CSV file",
		PrepareFiles: func(t *testing.T, dir string) {
			csvContent := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source`
			err := os.WriteFile(filepath.Join(dir, "empty.csv"), []byte(csvContent), 0644)
			require.NoError(t, err)
		},
		// No setup - converter should not be called since no videos found
		// Function returns nil early without creating output file
		ExpectedGames: nil,
	})

	validate(t, &testCase{
		Name: "Converter returns error",
		PrepareFiles: func(t *testing.T, dir string) {
			csvContent := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source
https://www.youtube.com/watch?v=video1,2023-01-01T12:00:00Z,Test Video 1,Test description,UCTEST123,video1,youtube`

			err := os.WriteFile(filepath.Join(dir, "UCTEST123.csv"), []byte(csvContent), 0644)
			require.NoError(t, err)
		},
		Setup: func(t *testing.T, converter *mockConverter.MockInterface) {
			converter.EXPECT().Convert(mock.AnythingOfType("[]*model.Video")).Return(nil, assert.AnError)
		},
		Error: "assert.AnError general error for testing",
	})

	validate(t, &testCase{
		Name: "Merge with existing data",
		PrepareFiles: func(t *testing.T, dir string) {
			csvContent := `Link,PublishedAt,Title,Description,ChannelID,VideoID,Source
https://www.youtube.com/watch?v=video1,2023-01-01T12:00:00Z,Test Video 1,Test description,UCTEST123,video1,youtube`

			err := os.WriteFile(filepath.Join(dir, "UCTEST123.csv"), []byte(csvContent), 0644)
			require.NoError(t, err)

			// Create existing output file
			outputPath := filepath.Join(dir, "output.json")
			existingGames := []*model.Game{
				{
					Name: "Existing Game",
					Content: []*model.Content{
						{
							Link:   "https://www.youtube.com/watch?v=existing",
							Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
							Source: model.SourceYouTube,
						},
					},
				},
			}

			file, err := os.Create(outputPath)
			require.NoError(t, err)
			defer func() { require.NoError(t, file.Close()) }()
			err = model.JSONWrite(file, existingGames)
			require.NoError(t, err)
		},
		Setup: func(t *testing.T, converter *mockConverter.MockInterface) {
			newGames := []*model.Game{
				{
					Name: "New Game",
					Content: []*model.Content{
						{
							Link:   "https://www.youtube.com/watch?v=video1",
							Start:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Source: model.SourceYouTube,
						},
					},
				},
			}

			converter.EXPECT().Convert(mock.AnythingOfType("[]*model.Video")).Return(newGames, nil)
		},
		ExpectedGames: []*model.Game{
			{
				Name: "Existing Game",
				Content: []*model.Content{
					{
						Link:   "https://www.youtube.com/watch?v=existing",
						Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						Source: model.SourceYouTube,
					},
				},
			},
			{
				Name: "New Game",
				Content: []*model.Content{
					{
						Link:   "https://www.youtube.com/watch?v=video1",
						Start:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Source: model.SourceYouTube,
					},
				},
			},
		},
	})
}

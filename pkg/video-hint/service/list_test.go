package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/seggga/postgres/pkg/video-hint/storage"
)

func TestGetVideosByCaption(t *testing.T) {
	cases := []struct {
		SearchPhrase   string
		ExpectedPhrase string
		ExpectedVideos []*storage.FoundVideo
		MockErr        error
		ExpectedErr    error
	}{
		{
			SearchPhrase:   "interesting stuff",
			ExpectedPhrase: "interesting stuff",
			ExpectedVideos: []*storage.FoundVideo{
				{
					Caption:  "some awsome and interesting stuff",
					URI:      "https://superstuff.com/videos/super-stuff/intetesting/super",
					Location: "/videos/temp/trash/super",
				},
			},
			MockErr:     nil,
			ExpectedErr: nil,
		},
		{
			SearchPhrase:   "INTERESTING stuff",
			ExpectedPhrase: "interesting stuff",
			ExpectedVideos: []*storage.FoundVideo{
				{
					Caption:  "some awsome and interesting stuff",
					URI:      "https://superstuff.com/videos/super-stuff/interesting/super",
					Location: "/videos/temp/trash/super",
				},
			},
			MockErr:     nil,
			ExpectedErr: nil,
		},
		{
			SearchPhrase:   "",
			ExpectedVideos: nil,
			ExpectedErr:    ErrIncorrectCaptionSubstring,
		},
		{
			SearchPhrase:   "interting",
			ExpectedPhrase: "interting",
			ExpectedVideos: nil,
			MockErr:        fmt.Errorf("some simple err"),
			ExpectedErr:    ErrDBRequestFailed,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("test case #%d", i), func(t *testing.T) {
			mock := &dbMock{
				t:              t,
				expectedPrefix: tc.ExpectedPhrase,
				expectedError:  tc.MockErr,
				videosToReturn: tc.ExpectedVideos,
			}
			actualFoundVideos, actualErr := GetVideosByCaption(mock, tc.SearchPhrase)
			if t.Failed() {
				return
			}
			if err := compareErrs(tc.ExpectedErr, actualErr); err != nil {
				t.Error(err)
				return
			}
			if len(actualFoundVideos) != len(tc.ExpectedVideos) {
				t.Errorf("expected viceos len %d, got %d", len(tc.ExpectedVideos), len(actualFoundVideos))
				return
			}
			for i, expectedVideo := range tc.ExpectedVideos {
				actualVideo := *actualFoundVideos[i]
				if *expectedVideo != actualVideo {
					t.Errorf("video %d: expected: %v, got: %v", i, *expectedVideo, actualVideo)
					return
				}
			}
		})
	}
}

func compareErrs(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if actualErr == nil {
		return fmt.Errorf("expected an error \"%v\", got nil", expectedErr)
	}
	if !errors.Is(actualErr, expectedErr) {
		return fmt.Errorf("expected error \"%v\" and actual error \"%v\" are different", expectedErr, actualErr)
	}
	return nil
}

type dbMock struct {
	t              *testing.T
	expectedPrefix string
	expectedError  error
	videosToReturn []*storage.FoundVideo
}

func (db *dbMock) GetVideosByCaption(ctx context.Context, captionSubstring string) ([]*storage.FoundVideo, error) {
	if captionSubstring != db.expectedPrefix {
		db.t.Errorf("error in DB mock: expected search phrase: %s, got: %s", db.expectedPrefix, captionSubstring)
		return nil, nil
	}
	return db.videosToReturn, db.expectedError
}

func (db *dbMock) Close() {}

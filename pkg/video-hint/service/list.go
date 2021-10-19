package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/seggga/postgres/pkg/video-hint/storage"
)

var (
	ErrIncorrectCaptionSubstring = fmt.Errorf("got an incorrect caption substring")
	ErrDBRequestFailed           = fmt.Errorf("a request to DB failed")
)

func GetVideosByCaption(db storage.DB, captionSubstring string) ([]*storage.FoundVideo, error) {
	if len(captionSubstring) == 0 {
		return nil, fmt.Errorf("%w: passed search phrase is empty", ErrIncorrectCaptionSubstring)
	}
	videos, err := db.GetVideosByCaption(context.Background(), strings.ToLower(captionSubstring))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get videos by caption substring: %v", ErrDBRequestFailed, err)
	}
	return videos, nil
}

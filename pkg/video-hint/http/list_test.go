package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/seggga/postgres/pkg/video-hint/storage"
)

func TestGetVideosByCaption(t *testing.T) {
	cases := []struct {
		CaptionSubstring  string
		ExpectedSubstring string
		MockErr           error
		ExpectedRespCode  int
	}{
		{
			CaptionSubstring:  "repudiandae",
			ExpectedSubstring: "repudiandae",
			ExpectedRespCode:  http.StatusOK,
		},
		{
			CaptionSubstring: "",
			ExpectedRespCode: http.StatusBadRequest,
		},
		{
			CaptionSubstring:  "alidd",
			ExpectedSubstring: "alidd",
			MockErr:           fmt.Errorf("some err"),
			ExpectedRespCode:  http.StatusInternalServerError,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case #%d", i), func(t *testing.T) {
			t.Logf("caption substring: %s, expected resp code: %d", tc.CaptionSubstring, tc.ExpectedRespCode)

			urlPath := fmt.Sprintf("/videos/%s", tc.CaptionSubstring)
			req, err := http.NewRequest("GET", urlPath, nil)
			if err != nil {
				t.Errorf("failed to create an http request: %v", err)
				return
			}
			req = req.WithContext(context.WithValue(req.Context(), storage.ContextKeyDB, &dbMock{
				t:                 t,
				expectedSubstring: tc.CaptionSubstring,
				expectedError:     tc.MockErr,
				videosToReturn:    nil,
			}))

			handler := mux.NewRouter()
			handler.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
				GetVideosByCaption(w, r, tc.CaptionSubstring)
			}).Methods("GET")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.ExpectedRespCode {
				t.Errorf("expected code: %d, got: %d", tc.ExpectedRespCode, rr.Code)
			}
		})
	}
}

type dbMock struct {
	t                 *testing.T
	expectedSubstring string
	expectedError     error
	videosToReturn    []*storage.FoundVideo
}

func (db *dbMock) GetVideosByCaption(ctx context.Context, searchPhrase string) ([]*storage.FoundVideo, error) {
	if searchPhrase != db.expectedSubstring {
		db.t.Errorf("error in DB mock: expected search phrase: %s, got: %s", db.expectedSubstring, searchPhrase)
		return nil, nil
	}
	return db.videosToReturn, db.expectedError
}

func (db *dbMock) Close() {}

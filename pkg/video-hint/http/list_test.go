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
		EmailPrefix      string
		ExpectedPrefix   string
		MockErr          error
		ExpectedRespCode int
	}{
		{
			EmailPrefix:      "alidd",
			ExpectedPrefix:   "alidd",
			ExpectedRespCode: http.StatusOK,
		},
		{
			EmailPrefix:      "ALidd",
			ExpectedPrefix:   "alidd",
			ExpectedRespCode: http.StatusOK,
		},
		{
			EmailPrefix:      "",
			ExpectedRespCode: http.StatusBadRequest,
		},
		{
			EmailPrefix:      "alidd",
			ExpectedPrefix:   "alidd",
			MockErr:          fmt.Errorf("some err"),
			ExpectedRespCode: http.StatusInternalServerError,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case #%d", i), func(t *testing.T) {
			t.Logf("prefix: %s, expected resp code: %d", tc.EmailPrefix, tc.ExpectedRespCode)

			urlPath := fmt.Sprintf("/phone/%s", tc.EmailPrefix)
			req, err := http.NewRequest("GET", urlPath, nil)
			if err != nil {
				t.Errorf("failed to create an http request: %v", err)
				return
			}
			req = req.WithContext(context.WithValue(req.Context(), storage.ContextKeyDB, &dbMock{
				t:              t,
				expectedPrefix: tc.ExpectedPrefix,
				expectedError:  tc.MockErr,
				phonesToReturn: nil,
			}))

			handler := mux.NewRouter()
			handler.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
				GetVideosByCaption(w, r, tc.EmailPrefix)
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
	t              *testing.T
	expectedPrefix string
	expectedError  error
	phonesToReturn []*storage.FoundVideo
}

func (db *dbMock) GetVideosByCaption(ctx context.Context, prefix string) ([]*storage.FoundVideo, error) {
	if prefix != db.expectedPrefix {
		db.t.Errorf("error in DB mock: expected email prefix: %s, got: %s", db.expectedPrefix, prefix)
		return nil, nil
	}
	return db.phonesToReturn, db.expectedError
}

func (db *dbMock) Close() {}

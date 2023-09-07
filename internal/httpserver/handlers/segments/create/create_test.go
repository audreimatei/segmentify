package create_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"segmentify/internal/httpserver/handlers/segments/create"
	"segmentify/internal/httpserver/handlers/segments/create/mocks"
	"segmentify/internal/lib/logger/handlers/slogdiscard"
	"segmentify/internal/models"
)

func TestCreateHandler(t *testing.T) {
	cases := []struct {
		name      string
		slug      string
		respCode  int
		respError string
		mockError error
	}{
		{
			name:     "Success",
			slug:     "SHINY_NEW_SEGMENT",
			respCode: http.StatusCreated,
		},
		{
			name:      "Empty Slug",
			slug:      "",
			respCode:  http.StatusUnprocessableEntity,
			respError: "field Slug is a required field",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			segmentCreatorMock := mocks.NewSegmentCreator(t)

			if tc.respError == "" || tc.mockError != nil {
				segmentCreatorMock.On("CreateSegment", models.Segment{Slug: tc.slug}).
					Return(models.Segment{Slug: tc.slug}, tc.mockError).
					Once()
			}

			handler := create.New(slogdiscard.NewDiscardLogger(), segmentCreatorMock)

			input := fmt.Sprintf(`{"slug": "%s"}`, tc.slug)

			req, err := http.NewRequest(http.MethodPost, "/segments", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.respCode, rr.Code)

			body := rr.Body.String()

			var resp models.Segment

			require.NoError(t, json.Unmarshal([]byte(body), &resp))
		})
	}
}

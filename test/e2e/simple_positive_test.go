package e2e

import (
	"net/http"
	"net/url"
	createSegment "segmentify/internal/httpserver/handlers/segments/create"
	updateUserSegments "segmentify/internal/httpserver/handlers/users/update"
	"segmentify/internal/models"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8081"
)

func TestSegmentifySimplePositive(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Creating A, B, C segments
	segments := []string{"A", "B", "C"}
	for _, segment := range segments {
		resp := e.POST("/segments").
			WithJSON(createSegment.Request{Slug: segment}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Keys().ContainsOnly("slug")
		resp.Value("slug").String().IsEqual(segment)
	}

	// Creating a user
	e.POST("/users").
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		Value("id").Number().IsEqual(1)

	// Adding segments A, B, C to a user
	req := updateUserSegments.Request{
		SegmentsToAdd: []models.SegmentToAdd{
			{Slug: segments[0]},
			{Slug: segments[1]},
			{Slug: segments[2]},
		},
		SegmentsToRemove: []models.SegmentToRemove{},
	}
	e.PATCH("/users/1/segments").
		WithJSON(req).
		Expect().
		Status(http.StatusNoContent)

	// Removing segment B from user
	req = updateUserSegments.Request{
		SegmentsToAdd:    []models.SegmentToAdd{},
		SegmentsToRemove: []models.SegmentToRemove{{Slug: segments[1]}},
	}
	e.PATCH("/users/1/segments").
		WithJSON(req).
		Expect().
		Status(http.StatusNoContent)

	// Getting user segments A, C
	resp := e.GET("/users/1/segments").
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Keys().ContainsOnly("id", "segments")
	resp.Value("id").Number().IsEqual(1)
	new_segments := []string{"A", "C"}
	resp.Value("segments").IsEqual(new_segments)
}

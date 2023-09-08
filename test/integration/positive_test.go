package integration

import (
	"context"
	"net/http"
	"net/url"
	getUserSegments "segmentify/internal/httpserver/handlers/users/get"
	updateUserSegments "segmentify/internal/httpserver/handlers/users/update"
	"segmentify/internal/models"
	"slices"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

const (
	host  = "0.0.0.0:8081"
	PGURL = "postgres://postgres:password@0.0.0.0:5432/segmentify_test"
)

func cleanDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, PGURL)
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "TRUNCATE users, segments, users_segments, users_segments_history")
	require.NoError(t, err)
}

func TestUpdateUserSegments(t *testing.T) {
	cleanDB(t)
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Creating A, B, C segments
	segments := []string{"A", "B", "C"}
	for _, segment := range segments {
		resp := e.POST("/segments").
			WithJSON(models.Segment{Slug: segment}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Keys().ContainsOnly("slug", "percent")
		resp.Value("slug").String().IsEqual(segment)
	}

	// Creating a user
	var userResp map[string]int64
	e.POST("/users").
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Decode(&userResp)

	userID, ok := userResp["id"]
	require.Equal(t, true, ok)

	// Adding segments A, B, C to a user
	req := updateUserSegments.Request{
		SegmentsToAdd: []models.SegmentToAdd{
			{Slug: segments[0]},
			{Slug: segments[1]},
			{Slug: segments[2]},
		},
		SegmentsToRemove: []models.SegmentToRemove{},
	}
	e.PATCH("/users/{id}/segments", userID).
		WithJSON(req).
		Expect().
		Status(http.StatusNoContent)

	// Removing segment B from user
	req = updateUserSegments.Request{
		SegmentsToAdd:    []models.SegmentToAdd{},
		SegmentsToRemove: []models.SegmentToRemove{{Slug: segments[1]}},
	}
	e.PATCH("/users/{id}/segments", userID).
		WithJSON(req).
		Expect().
		Status(http.StatusNoContent)

	// Getting user segments A, C
	resp := e.GET("/users/{id}/segments", userID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Keys().ContainsOnly("id", "segments")
	resp.Value("id").Number().IsEqual(userID)
	new_segments := []string{"A", "C"}
	resp.Value("segments").IsEqual(new_segments)
}

func TestCreateSegmentWithPercent(t *testing.T) {
	cleanDB(t)
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Setting constants
	const (
		usersCount  = 10
		percent     = 50
		segmentSlug = "WOW"
	)

	// Creating users
	var usersIDs [usersCount]int64
	for i := 0; i < usersCount; i++ {
		var userResp map[string]int64
		e.POST("/users").
			Expect().
			Status(http.StatusCreated).
			JSON().Object().Decode(&userResp)

		userID, ok := userResp["id"]
		require.Equal(t, true, ok)
		usersIDs[i] = userID
	}

	// Creating segment with percent
	e.POST("/segments").
		WithJSON(models.Segment{Slug: segmentSlug, Percent: percent}).
		Expect().
		Status(http.StatusCreated)

	// Counting number of users with WOW segment
	usersWithSegmentCount := 0
	for _, id := range usersIDs {
		var resp getUserSegments.Response
		e.GET("/users/{id}/segments", id).
			Expect().
			Status(http.StatusOK).
			JSON().Object().Decode(&resp)

		if slices.Contains(resp.Segments, "WOW") {
			usersWithSegmentCount++
		}
	}
	require.Equal(t, usersCount*percent/100, usersWithSegmentCount)
}

package models

import "time"

type Segment struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

type SegmentToAdd struct {
	Slug     string    `json:"slug" validate:"required"`
	ExpireAt time.Time `json:"expire_at"`
}

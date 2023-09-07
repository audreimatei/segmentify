package models

import "time"

type Segment struct {
	Slug    string `json:"slug" validate:"required"`
	Percent int64  `json:"percent" validate:"gte=0,lte=100"`
}

type SegmentToAdd struct {
	Slug     string    `json:"slug" validate:"required"`
	ExpireAt time.Time `json:"expire_at"`
}

type SegmentToRemove struct {
	Slug string `json:"slug" validate:"required"`
}

package models

import "time"

type Segment struct {
	Slug    string `json:"slug" validate:"required"`
	Percent int64  `json:"percent" validate:"gte=0,lte=100"`
}

type SegmentToAdd struct {
	Slug     string    `json:"slug" validate:"required"`
	ExpireAt time.Time `json:"expire_at" example:"2023-09-12T15:49:26Z"`
}

type SegmentToRemove struct {
	Slug string `json:"slug" validate:"required"`
}

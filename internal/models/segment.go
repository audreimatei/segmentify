package models

import "time"

type SegmentToAdd struct {
	Slug     string    `json:"slug" validate:"required"`
	ExpireAt time.Time `json:"expire_at"`
}

package storage

import (
	"fmt"
)

type ErrSegmentNotFound struct {
	Slug string
}

func (e ErrSegmentNotFound) Error() string {
	return fmt.Sprintf("segment with slug=%s not found", e.Slug)
}

type ErrSegmentExists struct {
	Slug string
}

func (e ErrSegmentExists) Error() string {
	return fmt.Sprintf("segment with slug=%s exists", e.Slug)
}

type ErrUserNotFound struct {
	ID int64
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with id=%d not found", e.ID)
}

type ErrUserSegmentNotFound struct {
	Slug string
}

func (e ErrUserSegmentNotFound) Error() string {
	return fmt.Sprintf("user segment with slug=%s not found", e.Slug)
}

type ErrUserSegmentExists struct {
	Slug string
}

func (e ErrUserSegmentExists) Error() string {
	return fmt.Sprintf("user segment with slug=%s exists", e.Slug)
}

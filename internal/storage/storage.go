package storage

import "errors"

var (
	ErrSegmentNotFound     = errors.New("segment not found")
	ErrSegmentExists       = errors.New("segment exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserSegmentNotFound = errors.New("user segment not found")
	ErrUserSegmentExists   = errors.New("user segment exits")
)

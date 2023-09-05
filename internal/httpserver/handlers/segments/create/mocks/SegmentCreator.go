// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SegmentCreator is an autogenerated mock type for the SegmentCreator type
type SegmentCreator struct {
	mock.Mock
}

// CreateSegment provides a mock function with given fields: slug
func (_m *SegmentCreator) CreateSegment(slug string) (string, error) {
	ret := _m.Called(slug)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(slug)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(slug)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(slug)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewSegmentCreator creates a new instance of SegmentCreator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSegmentCreator(t interface {
	mock.TestingT
	Cleanup(func())
}) *SegmentCreator {
	mock := &SegmentCreator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

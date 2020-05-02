// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	parser "github.com/darkclainer/camgo/pkg/parser"
	mock "github.com/stretchr/testify/mock"
)

// QueryInterface is an autogenerated mock type for the QueryInterface type
type QueryInterface struct {
	mock.Mock
}

// Close provides a mock function with given fields: ctx
func (_m *QueryInterface) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLemma provides a mock function with given fields: ctx, lemmaID
func (_m *QueryInterface) GetLemma(ctx context.Context, lemmaID string) ([]*parser.Lemma, error) {
	ret := _m.Called(ctx, lemmaID)

	var r0 []*parser.Lemma
	if rf, ok := ret.Get(0).(func(context.Context, string) []*parser.Lemma); ok {
		r0 = rf(ctx, lemmaID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*parser.Lemma)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, lemmaID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Search provides a mock function with given fields: ctx, query
func (_m *QueryInterface) Search(ctx context.Context, query string) (string, []string, error) {
	ret := _m.Called(ctx, query)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, query)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 []string
	if rf, ok := ret.Get(1).(func(context.Context, string) []string); ok {
		r1 = rf(ctx, query)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]string)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

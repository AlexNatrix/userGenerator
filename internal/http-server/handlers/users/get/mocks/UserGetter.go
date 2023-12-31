// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	internal "usergenerator/internal/lib/api/model/user"

	mock "github.com/stretchr/testify/mock"
)

// UserGetter is an autogenerated mock type for the UserGetter type
type UserGetter struct {
	mock.Mock
}

// GetUsers provides a mock function with given fields: userQuery
func (_m *UserGetter) GetUsers(userQuery map[string][]string) ([]internal.User, error) {
	ret := _m.Called(userQuery)

	var r0 []internal.User
	var r1 error
	if rf, ok := ret.Get(0).(func(map[string][]string) ([]internal.User, error)); ok {
		return rf(userQuery)
	}
	if rf, ok := ret.Get(0).(func(map[string][]string) []internal.User); ok {
		r0 = rf(userQuery)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]internal.User)
		}
	}

	if rf, ok := ret.Get(1).(func(map[string][]string) error); ok {
		r1 = rf(userQuery)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewUserGetter interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserGetter creates a new instance of UserGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserGetter(t mockConstructorTestingTNewUserGetter) *UserGetter {
	mock := &UserGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

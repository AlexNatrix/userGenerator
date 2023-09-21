// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	internal "usergenerator/internal/lib/api/model/user"

	mock "github.com/stretchr/testify/mock"
)

// UserUpdater is an autogenerated mock type for the UserUpdater type
type UserUpdater struct {
	mock.Mock
}

// UpdateUser provides a mock function with given fields: userID, user
func (_m *UserUpdater) UpdateUser(userID int64, user internal.User) error {
	ret := _m.Called(userID, user)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, internal.User) error); ok {
		r0 = rf(userID, user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewUserUpdater interface {
	mock.TestingT
	Cleanup(func())
}

// NewUserUpdater creates a new instance of UserUpdater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUserUpdater(t mockConstructorTestingTNewUserUpdater) *UserUpdater {
	mock := &UserUpdater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

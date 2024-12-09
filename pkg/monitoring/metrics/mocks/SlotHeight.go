// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	types "github.com/goplugin/plugin-solana/pkg/monitoring/types"
	mock "github.com/stretchr/testify/mock"
)

// SlotHeight is an autogenerated mock type for the SlotHeight type
type SlotHeight struct {
	mock.Mock
}

// Cleanup provides a mock function with given fields:
func (_m *SlotHeight) Cleanup() {
	_m.Called()
}

// Set provides a mock function with given fields: slot, chain, url
func (_m *SlotHeight) Set(slot types.SlotHeight, chain string, url string) {
	_m.Called(slot, chain, url)
}

// NewSlotHeight creates a new instance of SlotHeight. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSlotHeight(t interface {
	mock.TestingT
	Cleanup(func())
}) *SlotHeight {
	mock := &SlotHeight{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
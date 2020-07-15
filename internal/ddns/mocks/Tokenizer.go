// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	ddns "github.com/xsteadfastx/carson/internal/ddns"
)

// Tokenizer is an autogenerated mock type for the Tokenizer type
type Tokenizer struct {
	mock.Mock
}

// Create provides a mock function with given fields: secret, hostname, recordType
func (_m *Tokenizer) Create(secret string, hostname string, recordType string) (string, error) {
	ret := _m.Called(secret, hostname, recordType)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string) string); ok {
		r0 = rf(secret, hostname, recordType)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(secret, hostname, recordType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Parse provides a mock function with given fields: secret, token
func (_m *Tokenizer) Parse(secret string, token string) (ddns.Record, error) {
	ret := _m.Called(secret, token)

	var r0 ddns.Record
	if rf, ok := ret.Get(0).(func(string, string) ddns.Record); ok {
		r0 = rf(secret, token)
	} else {
		r0 = ret.Get(0).(ddns.Record)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(secret, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
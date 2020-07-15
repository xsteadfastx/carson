// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	ddns "github.com/xsteadfastx/carson/internal/ddns"
)

// ZoneRefresher is an autogenerated mock type for the ZoneRefresher type
type ZoneRefresher struct {
	mock.Mock
}

// Refresh provides a mock function with given fields: ctx, _a1, r, remoteAddr
func (_m *ZoneRefresher) Refresh(ctx context.Context, _a1 *ddns.DDNS, r ddns.Record, remoteAddr string) {
	_m.Called(ctx, _a1, r, remoteAddr)
}
package ddns

import (
	"context"
)

func GetUUID(ctx context.Context) string {
	switch v := ctx.Value(ReqID).(type) {
	case string:
		return v
	default:
		return "none"
	}
}

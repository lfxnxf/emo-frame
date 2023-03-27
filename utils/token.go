package utils

import "context"

const (
	TokenKey = "token"
	UidKey   = "uid"
)

func GetToken(ctx context.Context) string {
	token, ok := ctx.Value(TokenKey).(string)
	if !ok {
		return ""
	}
	return token
}

func GetUid(ctx context.Context) int64 {
	uid, ok := ctx.Value(UidKey).(int64)
	if !ok {
		return 0
	}
	return uid
}

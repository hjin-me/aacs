package middlewares

import (
	"context"
	"errors"
)

var ErrNoSession = errors.New("没有session")

func GetUID(ctx context.Context) (string, error) {
	uid, _, err := GetSession(ctx)
	return uid, err
}
func GetToken(ctx context.Context) (string, error) {
	_, tk, err := GetSession(ctx)
	return tk, err
}
func GetSession(ctx context.Context) (uid, token string, err error) {
	sub, ok := FromContext(ctx)
	if !ok {
		return "", "", ErrNoSession
	}
	uid = sub.UID
	token = sub.Token
	if uid == "" {
		return "", "", ErrNoSession
	}
	if token == "" {
		return "", "", ErrNoSession
	}
	return uid, token, nil
}

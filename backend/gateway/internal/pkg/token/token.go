package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

type TokenExtractor interface {
	ExtractTTL(accessToken string) (time.Duration, error)
	ExtractEmail(accessToken string) (string, error)
}

type AccessTokenExtractor struct{}

func extractPayload(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var payloadData map[string]interface{}
	if err = json.Unmarshal(payload, &payloadData); err != nil {
		return nil, err
	}

	return payloadData, nil
}

func (e *AccessTokenExtractor) ExtractTTL(token string) (ttl time.Duration, err error) {
	payloadData, err := extractPayload(token)
	if err != nil {
		return 0, err
	}

	exp, ok := payloadData["exp"].(float64)
	if !ok {
		return 0, errors.New("exp not found in the token")
	}

	expString := strconv.FormatFloat(exp, 'f', -1, 64)
	i, err := strconv.ParseInt(expString, 10, 64)
	if err != nil {
		return 0, err
	}
	tExp := time.Unix(i, 0)

	return time.Until(tExp), nil
}

func (e *AccessTokenExtractor) ExtractEmail(token string) (email string, err error) {
	payloadData, err := extractPayload(token)
	if err != nil {
		return "", err
	}

	email, ok := payloadData["email"].(string)
	if !ok {
		return "", errors.New("email not found in the token")
	}

	return email, nil
}

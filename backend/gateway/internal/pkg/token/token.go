package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

func ExtractTTL(token string) (ttl time.Duration, err error) {
	defer func() {
		if r := recover(); r != nil {
			ttl = 0
			err = errors.New("failed to get ttl")
		}
	}()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, errors.New("cannot split the token string")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, err
	}

	var payloadData map[string]interface{}
	if err = json.Unmarshal(payload, &payloadData); err != nil {
		return 0, err
	}

	exp, ok := payloadData["exp"]
	if !ok {
		return 0, errors.New("exp not found in the token")
	}

	expString := strconv.FormatFloat(exp.(float64), 'f', -1, 64)
	i, err := strconv.ParseInt(expString, 10, 64)
	if err != nil {
		return 0, err
	}
	tExp := time.Unix(i, 0)

	return time.Until(tExp), nil
}

func ExtractEmail(token string) (email string, err error) {
	defer func() {
		if r := recover(); r != nil {
			email = ""
			err = errors.New("failed to get ttl")
		}
	}()

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("cannot split the token string")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var payloadData map[string]interface{}
	if err = json.Unmarshal(payload, &payloadData); err != nil {
		return "", err
	}

	email, ok := payloadData["email"].(string)
	if !ok {
		return "", errors.New("exp not found in the token")
	}

	return email, nil
}

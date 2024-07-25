package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
	"strings"
	"time"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

func getTTL(token string) (ttl time.Duration, err error) {
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

func getEmail(token string) (email string, err error) {
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

type user struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) UserRegister(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO add validate method with regex (for example)
	if u.Email == "" || u.Password == "" {
		h.logger.Error("data is not valid", zap.String("email", u.Email), zap.String("password", u.Password))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Register(context.TODO(), &pb.RegisterRequest{
		Email:    u.Email,
		Password: u.Password,
	})
	if err != nil {
		h.logger.Error("show error", zap.Error(err))
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed auth request",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)

			if e.Code() == codes.AlreadyExists {
				http.Error(w, "User already exist", http.StatusConflict)
				return
			}

			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		} else {
			h.logger.Error("unexpected error", zap.Error(err))
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ttl, err := getTTL(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to get ttl", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rRes := h.redis.Client.Set("access_token:"+res.AccessToken, u.Email, ttl)
	if rRes.Err() != nil {
		h.logger.Error("failed to set access token", zap.String("token", res.AccessToken), zap.Error(rRes.Err()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:   "refresh_token",
		Value:  res.RefreshToken,
		Path:   "/",
		MaxAge: 604800,
		//Secure:   true,
		//HttpOnly: true,
		SameSite: 0,
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UserLogin(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO add validate method with regex (for example)
	if u.Email == "" || u.Password == "" {
		h.logger.Error("data is not valid", zap.String("email", u.Email), zap.String("password", u.Password))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Login(context.TODO(), &pb.LoginRequest{
		Email:    u.Email,
		Password: u.Password,
	})
	if err != nil {
		h.logger.Error("show error", zap.Error(err))
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed auth request",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		} else {
			h.logger.Error("unexpected error", zap.Error(err))
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ttl, err := getTTL(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to get ttl", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rRes := h.redis.Client.Set("access_token:"+res.AccessToken, u.Email, ttl)
	if rRes.Err() != nil {
		h.logger.Error("failed to set access token", zap.String("token", res.AccessToken), zap.Error(rRes.Err()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/",
		MaxAge:   604800,
		Secure:   true,
		HttpOnly: true,
		SameSite: 0,
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UserRefreshToken(w http.ResponseWriter, r *http.Request) {
	rt, err := r.Cookie("refresh_token")
	if err != nil {
		h.logger.Error("no refresh token cookies", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Debug("refresh token", zap.Any("value", rt.Value), zap.Any("MaxAge", rt.MaxAge), zap.Any("HttpOnly", rt.HttpOnly))

	res, err := h.auth.RefreshToken(context.TODO(), &pb.RefreshTokenRequest{
		RefreshToken: rt.Value,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed to refresh token",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		} else {
			h.logger.Error("unexpected error", zap.Error(err))
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ttl, err := getTTL(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to get ttl", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email, err := getEmail(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to get email", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rRes := h.redis.Client.Set("access_token:"+res.AccessToken, email, ttl)
	if rRes.Err() != nil {
		h.logger.Error("failed to set access token", zap.String("token", res.AccessToken), zap.Error(rRes.Err()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusOK)
}

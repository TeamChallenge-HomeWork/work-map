package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/pkg/token"
)

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

	ttl, err := token.ExtractTTL(res.AccessToken)
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

	ttl, err := token.ExtractTTL(res.AccessToken)
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

	ttl, err := token.ExtractTTL(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to get ttl", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email, err := token.ExtractEmail(res.AccessToken)
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

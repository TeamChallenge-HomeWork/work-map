package handlers

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

type accessTokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
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
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		} else {
			h.logger.Error("unexpected error", zap.Error(err))
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// TODO add JWT validator for access and refresh token

	// TODO add TTL extractor from JWT
	rRes := h.redis.Client.Set("access_token:"+res.AccessToken, u.Email, time.Duration(5)*time.Minute)
	if rRes.Err() != nil {
		h.logger.Error("failed to set access token", zap.String("token", res.AccessToken))
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
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(accessTokenResponse{
		AccessToken: res.AccessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

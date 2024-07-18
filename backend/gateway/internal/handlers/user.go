package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
	"strings"
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
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(accessTokenResponse{
		AccessToken: res.AccessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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

	return tExp.Sub(time.Now()), nil
}

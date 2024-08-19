package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
	"time"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/models"
	"workmap/gateway/internal/pkg/token"
)

func (h *Handler) UserRegister(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := u.Validate(); err != nil {
		h.logger.Error("user data is not valid", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Register(context.TODO(), &pb.RegisterRequest{
		Email:    u.Email,
		Password: u.Password,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed auth request",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)

			if e.Code() == codes.AlreadyExists {
				http.Error(w, "User email taken", http.StatusConflict)
				return
			}

			http.Error(w, e.Proto().Message, http.StatusBadRequest)
			return
		}

		h.logger.Error("unexpected error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.tokenStore.SaveAccessToken(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to save access token to redis store", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	e := &token.AccessTokenExtractor{}
	rTtl, err := e.ExtractTTL(res.RefreshToken)
	if err != nil {
		h.logger.Error("failed to get ttl from refresh token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	exp := time.Now().Add(rTtl)

	cookie := &http.Cookie{
		Name:  "refresh_token",
		Value: res.RefreshToken,
		Path:  "/",
		//Secure:   true,
		//HttpOnly: true,
		SameSite: 0,
		Expires:  exp,
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusCreated)

	h.logger.Info("user register success", zap.String("email", u.Email))
}

func (h *Handler) UserLogin(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := u.Validate(); err != nil {
		h.logger.Error("user data is not valid", zap.Error(err))
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
			http.Error(w, e.Proto().Message, http.StatusBadRequest)
			return
		}

		h.logger.Error("unexpected error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.tokenStore.SaveAccessToken(res.AccessToken)
	if err != nil {
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

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusOK)

	h.logger.Info("user login success", zap.String("email", u.Email))
}

func (h *Handler) UserRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.logger.Error("no refresh token cookies", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rt := cookie.Value

	res, err := h.auth.RefreshToken(context.TODO(), &pb.RefreshTokenRequest{
		RefreshToken: rt,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed to refresh token",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)
			http.Error(w, e.Proto().Message, http.StatusBadRequest)
			return
		}

		h.logger.Error("unexpected error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.tokenStore.SaveAccessToken(res.AccessToken)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", res.AccessToken))
	w.WriteHeader(http.StatusOK)

	e := &token.AccessTokenExtractor{}
	email, err := e.ExtractEmail(res.AccessToken)
	if err != nil {
		h.logger.Error("failed to extract email from access token", zap.String("refresh token", rt))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.logger.Info("user refresh token success", zap.String("email", email))
}

func (h *Handler) UserLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.logger.Error("no refresh token cookies", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rt := cookie.Value

	at := r.Header.Get("Authorization")
	if at == "" {
		h.logger.Error("no access token in header", zap.Error(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	} else {
		at = strings.Replace(at, "Bearer ", "", -1)
	}

	e := &token.AccessTokenExtractor{}
	email, err := e.ExtractEmail(at)
	if err != nil {
		h.logger.Error("failed to extract email from access token", zap.String("access token", at))
		http.Error(w, "Wrong auth token", http.StatusBadRequest)
		return
	}

	res, err := h.auth.Logout(context.TODO(), &pb.LogoutRequest{
		RefreshToken: rt,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			h.logger.Error(
				"failed to refresh token",
				zap.String("code", e.Code().String()),
				zap.String("description", e.Proto().Message),
			)
			http.Error(w, e.Proto().Message, http.StatusBadRequest)
			return
		} else {
			h.logger.Error("unexpected error", zap.Error(err))
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if res.IsSuccess {
		err = h.tokenStore.DeleteAccessToken(at)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Authorization", "")
	w.WriteHeader(http.StatusOK)

	h.logger.Info("user logout success", zap.String("email", email))
}

func (h *Handler) UserProfile(w http.ResponseWriter, r *http.Request) {
	bearer := r.Header.Get("Authorization")
	at := strings.TrimPrefix(bearer, "Bearer ")

	e := &token.AccessTokenExtractor{}
	email, err := e.ExtractEmail(at)
	if err != nil {
		fmt.Println(err)
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(struct {
		Email string `json:"email"`
	}{
		Email: email,
	})
	if err != nil {
		fmt.Println("merr", err)
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.logger.Error(err.Error())
	}
}

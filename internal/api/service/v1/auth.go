package v1

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dmytro-vovk/tro/internal/api/model"
)

const (
	salt       = "kbw4Gfd45DdBdf35tsMic14"
	signingKey = "A24njRJ4bUks9DmRfs23FDp"
	tokenTTL   = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

func (s *service) CreateUser(user model.User) (int, error) {
	user.Password = s.generatePasswordHash(user.Password)

	return s.db.CreateUser(user)
}

func (s *service) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func (s *service) GenerateToken(username, password string) (string, error) {
	user, err := s.db.GetUser(username, s.generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(tokenTTL).Unix(),
			IssuedAt:  now.Unix(),
		},
		UserID: user.ID,
	})

	return token.SignedString([]byte(signingKey))
}

func (s *service) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims aren't of type")
	}

	return claims.UserID, nil
}

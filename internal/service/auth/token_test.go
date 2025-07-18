package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseToken_Success(t *testing.T) {
	tm := auth.NewTokenManager("testsecret", time.Minute)

	tokenStr, err := tm.GenerateToken(42)
	assert.NoError(t, err)

	userID, err := tm.ParseToken(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), userID)
}

func TestParseToken_InvalidSignature(t *testing.T) {
	
	tm := auth.NewTokenManager("secretA", time.Minute)
	tokenStr, err := tm.GenerateToken(1)
	assert.NoError(t, err)

	
	tm2 := auth.NewTokenManager("secretB", time.Minute)
	_, err = tm2.ParseToken(tokenStr)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseToken_InvalidFormat(t *testing.T) {
	tm := auth.NewTokenManager("whatever", time.Minute)

	_, err := tm.ParseToken("not.a.valid.jwt")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseToken_Expired(t *testing.T) {
	tm := auth.NewTokenManager("secret", -time.Second) 

	tokenStr, err := tm.GenerateToken(123)
	assert.NoError(t, err)

	_, err = tm.ParseToken(tokenStr)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseToken_MissingUserIDClaim(t *testing.T) {
	
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("secret"))

	tm := auth.NewTokenManager("secret", time.Minute)
	_, err := tm.ParseToken(tokenStr)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"sso/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    123,
		Email: "test@example.com",
	}

	app := models.App{
		ID:     456,
		Secret: "test-secret",
	}

	duration := time.Hour

	before := time.Now()
	tokenString, err := NewToken(user, app, duration)
	after := time.Now()

	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token string, got empty string")
	}

	parsedToken, err := jwtlib.Parse(tokenString, func(_ *jwtlib.Token) (interface{}, error) {
		return []byte(app.Secret), nil
	})

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Fatal("expected token to be valid")
	}

	claims, ok := parsedToken.Claims.(jwtlib.MapClaims)
	if !ok {
		t.Fatal("expected jwt.MapClaims")
	}

	if got := claims["email"]; got != user.Email {
		t.Fatalf("expected email %q, got %v", user.Email, got)
	}

	if got := claims["uid"]; got != float64(user.ID) {
		t.Fatalf("expected uid %v, got %v", user.ID, got)
	}

	if got := claims["app_id"]; got != float64(app.ID) {
		t.Fatalf("expected app_id %v, got %v", app.ID, got)
	}

	expRaw, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("expected exp to be float64, got %T", claims["exp"])
	}

	exp := int64(expRaw)

	minExp := before.Add(duration).Unix()
	maxExp := after.Add(duration).Unix()

	if exp < minExp || exp > maxExp {
		t.Fatalf("expected exp between %d and %d, got %d", minExp, maxExp, exp)
	}
}

func TestNewToken_InvalidSecret(t *testing.T) {
	user := models.User{
		ID:    123,
		Email: "test@example.com",
	}

	app := models.App{
		ID:     456,
		Secret: "test-secret",
	}

	tokenString, err := NewToken(user, app, time.Hour)
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}

	parsedToken, err := jwtlib.Parse(tokenString, func(_ *jwtlib.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})

	if err == nil {
		t.Fatal("expected error with wrong secret, got nil")
	}

	if parsedToken != nil && parsedToken.Valid {
		t.Fatal("expected token to be invalid with wrong secret")
	}
}

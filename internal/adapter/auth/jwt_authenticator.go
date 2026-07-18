package auth

import (
	"errors"
	"fmt"
	"gin-api-gateway/internal/domain/model"
	"gin-api-gateway/internal/domain/service"

	"github.com/golang-jwt/jwt/v5"
)

type jwtAuthenticator struct {
	secretKey string
}

// NewJWTAuthenticator menginisialisasi instansi baru dari JWT Authenticator.
// NewJWTAuthenticator initializes a new instance of JWT Authenticator.
//
// Parameters:
//   - secretKey (string): Kunci rahasia untuk memvalidasi tanda tangan JWT / Secret key to validate JWT signature.
//
// Returns:
//   - service.Authenticator: Implementasi antarmuka Authenticator / Authenticator interface implementation.
func NewJWTAuthenticator(secretKey string) service.Authenticator {
	return &jwtAuthenticator{secretKey: secretKey}
}

// Verify memvalidasi token JWT dan mengekstrak klaim pengguna (user_id dan role).
// Verify validates the JWT token and extracts user claims (user_id and role).
//
// Parameters:
//   - tokenStr (string): Token JWT mentah dari request header / Raw JWT token string from request header.
//
// Returns:
//   - *model.UserClaims: Data klaim pengguna / User claims entity.
//   - error: Error jika token kedaluwarsa, tidak valid, atau salah format / Error if token is expired, invalid, or malformed.
func (a *jwtAuthenticator) Verify(tokenStr string) (*model.UserClaims, error) {
	// 1. Mengurai (parse) string token JWT dan memvalidasi algoritma enkripsi tanda tangan (signing method)
	// 1. Parse the JWT token string and validate the signature encryption algorithm (signing method)
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token encryption method is not valid")
		}
		return []byte(a.secretKey), nil
	})

	// 2. Memeriksa apakah terjadi kesalahan selama proses penguraian token
	// 2. Check if an error occurred during the token parsing process
	if err != nil {
		return nil, err
	}

	// 3. Memvalidasi keaktifan dan keaslian token
	// 3. Validate token expiration and authenticity
	if !token.Valid {
		return nil, errors.New("token is invalid or has expired")
	}

	// 4. Mengekstrak data klaim (claims) dari payload token
	// 4. Extract claims data from the token payload
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse token claims")
	}

	// 5. Mengambil informasi user_id dan role dari objek klaim
	// 5. Retrieve user_id and role information from the claims object
	userID, okUserID := claims["user_id"]
	role, okRole := claims["role"]
	if !okUserID || !okRole {
		return nil, errors.New("token claims are missing user_id or role")
	}

	// 6. Mengembalikan data klaim pengguna yang telah berhasil divalidasi ke dalam format model domain
	// 6. Return the successfully validated user claims mapped into domain model format
	return &model.UserClaims{
		UserID: fmt.Sprintf("%v", userID),
		Role:   fmt.Sprintf("%v", role),
	}, nil
}

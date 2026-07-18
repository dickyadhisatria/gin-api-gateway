package model

// UserClaims merepresentasikan data klaim identitas pengguna yang diekstrak dari token JWT.
// UserClaims represents the user identity claims extracted from the JWT token.
type UserClaims struct {
	UserID string // ID unik pengguna / Unique user ID
	Role   string // Peran pengguna (misalnya: admin, user) / User role (e.g. admin, user)
}

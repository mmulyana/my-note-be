package users

func ToProfileResponse(u *User) ProfileResponse {
	return ProfileResponse{
		ID:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		Photo:    u.Photo,
		IsGuest:  u.IsGuest,
	}
}

func ToTokenResponse(u *User, accessToken, refreshToken string, expiresAt int64) TokenResponse {
	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Email:        u.EmailOrEmpty(),
		IsGuest:      u.IsGuest,
	}
}

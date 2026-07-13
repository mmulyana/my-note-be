package users

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresAt   int64  `json:"expiresAt"`
	Email       string `json:"email"`
}

type UpdateProfileInput struct {
	Username *string `json:"username"`
	Photo    *string `json:"photo"`
}

type ProfileResponse struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	Username *string `json:"username"`
	Photo    *string `json:"photo"`
}

func toProfileResponse(u *User) ProfileResponse {
	return ProfileResponse{
		ID:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		Photo:    u.Photo,
	}
}

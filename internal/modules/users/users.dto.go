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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	Email        string `json:"email,omitempty"`
	IsGuest      bool   `json:"isGuest"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

type UpdateProfileInput struct {
	Username *string `json:"username"`
	Photo    *string `json:"photo"`
}

type ProfileResponse struct {
	ID       string  `json:"id"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Photo    *string `json:"photo"`
	IsGuest  bool    `json:"isGuest"`
}

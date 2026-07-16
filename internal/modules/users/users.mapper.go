package users

func ToProfileResponse(u *User) ProfileResponse {
	return ProfileResponse{
		ID:       u.ID.String(),
		Email:    u.Email,
		Username: u.Username,
		Photo:    u.Photo,
	}
}

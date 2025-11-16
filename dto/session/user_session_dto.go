package dto_session

type UserDto struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type UserSessionDto struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	User         UserDto `json:"user"`
	RefreshToken string  `json:"refresh_token"`
}

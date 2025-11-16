package dto_session

import "time"

type TokenResponseDTO struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

package auth

import (
	"fmt"
	"time"

	"firstpersoncode/go-uploader/domain"
	dto_session "firstpersoncode/go-uploader/dto/session"
	"firstpersoncode/go-uploader/internal/util"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

func NewAuthService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) domain.SessionService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *authService) RegisterUser(credentials *dto_session.TokenRequestDTO) error {
	if credentials.Username == "" || credentials.Password == "" {
		return fmt.Errorf("username and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	user := &domain.User{
		Username: credentials.Username,
		Password: string(hashedPassword),
	}

	if _, err := s.userRepo.Save(user); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (s *authService) CreateSession(credentials *dto_session.TokenRequestDTO) (*dto_session.TokenResponseDTO, error) {
	if credentials.Username == "" || credentials.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	user, err := s.userRepo.FindByUsername(credentials.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	session := &domain.Session{
		UserID: user.ID,
	}

	newSession, err := s.sessionRepo.Save(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	token, err := util.GenerateJWT(newSession.ID, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &dto_session.TokenResponseDTO{
		Token:  token,
		Expiry: expiry,
	}, nil
}

func (s *authService) GetUserSession(session *domain.Session) (*dto_session.UserSessionDto, error) {
	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return &dto_session.UserSessionDto{
		ID:     session.ID,
		UserID: session.UserID,
		User: dto_session.UserDto{
			ID:       user.ID,
			Username: user.Username,
		},
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*dto_session.TokenResponseDTO, error) {
	return nil, fmt.Errorf("not implemented")
}

package auth

import (
	"testing"
	"time"

	"firstpersoncode/go-uploader/domain"
	dto_session "firstpersoncode/go-uploader/dto/session"
	"firstpersoncode/go-uploader/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

func setupTestService() (domain.SessionService, domain.UserRepository, domain.SessionRepository) {
	userRepo := repositories.NewUserRepository()
	sessionRepo := repositories.NewSessionRepository()
	service := NewAuthService(userRepo, sessionRepo)
	return service, userRepo, sessionRepo
}

func TestRegisterUser_Success(t *testing.T) {
	service, userRepo, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}

	err := service.RegisterUser(credentials)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify user was created
	user, err := userRepo.FindByUsername("testuser")
	if err != nil {
		t.Fatalf("expected user to be created, got error: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", user.Username)
	}

	// Verify password was hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
	if err != nil {
		t.Error("password was not properly hashed")
	}
}

func TestRegisterUser_EmptyUsername(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "",
		Password: "password123",
	}

	err := service.RegisterUser(credentials)
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}

	if err.Error() != "username and password are required" {
		t.Errorf("expected 'username and password are required' error, got %v", err)
	}
}

func TestRegisterUser_EmptyPassword(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "",
	}

	err := service.RegisterUser(credentials)
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}

	if err.Error() != "username and password are required" {
		t.Errorf("expected 'username and password are required' error, got %v", err)
	}
}

func TestRegisterUser_DuplicateUsername(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}

	// Register first user
	err := service.RegisterUser(credentials)
	if err != nil {
		t.Fatalf("expected no error on first registration, got %v", err)
	}

	// Try to register duplicate
	err = service.RegisterUser(credentials)
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
}

func TestCreateSession_Success(t *testing.T) {
	service, _, _ := setupTestService()

	// First register a user
	registerCredentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}
	err := service.RegisterUser(registerCredentials)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Now create a session
	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}

	tokenResponse, err := service.CreateSession(credentials)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokenResponse.Token == "" {
		t.Error("expected token to be generated")
	}

	if tokenResponse.Expiry.Before(time.Now()) {
		t.Error("expected expiry to be in the future")
	}

	// Verify expiry is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := tokenResponse.Expiry.Sub(expectedExpiry)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("expected expiry to be ~24 hours from now, got %v", tokenResponse.Expiry)
	}
}

func TestCreateSession_EmptyUsername(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "",
		Password: "password123",
	}

	_, err := service.CreateSession(credentials)
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}

	if err.Error() != "username and password are required" {
		t.Errorf("expected 'username and password are required' error, got %v", err)
	}
}

func TestCreateSession_EmptyPassword(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "",
	}

	_, err := service.CreateSession(credentials)
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}

	if err.Error() != "username and password are required" {
		t.Errorf("expected 'username and password are required' error, got %v", err)
	}
}

func TestCreateSession_InvalidUsername(t *testing.T) {
	service, _, _ := setupTestService()

	credentials := &dto_session.TokenRequestDTO{
		Username: "nonexistent",
		Password: "password123",
	}

	_, err := service.CreateSession(credentials)
	if err == nil {
		t.Fatal("expected error for invalid username, got nil")
	}

	if err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials' error, got %v", err)
	}
}

func TestCreateSession_InvalidPassword(t *testing.T) {
	service, _, _ := setupTestService()

	// First register a user
	registerCredentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}
	err := service.RegisterUser(registerCredentials)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Try to login with wrong password
	credentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = service.CreateSession(credentials)
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}

	if err.Error() != "invalid credentials" {
		t.Errorf("expected 'invalid credentials' error, got %v", err)
	}
}

func TestGetUserSession_Success(t *testing.T) {
	service, userRepo, sessionRepo := setupTestService()

	// Register and create session
	registerCredentials := &dto_session.TokenRequestDTO{
		Username: "testuser",
		Password: "password123",
	}
	err := service.RegisterUser(registerCredentials)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	user, _ := userRepo.FindByUsername("testuser")

	// Create a session manually
	session := &domain.Session{
		UserID: user.ID,
	}
	savedSession, err := sessionRepo.Save(session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Get user session
	userSessionDto, err := service.GetUserSession(savedSession)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userSessionDto.ID != savedSession.ID {
		t.Errorf("expected session ID %s, got %s", savedSession.ID, userSessionDto.ID)
	}

	if userSessionDto.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, userSessionDto.UserID)
	}

	if userSessionDto.User.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", userSessionDto.User.Username)
	}

	if userSessionDto.RefreshToken != savedSession.RefreshToken {
		t.Errorf("expected refresh token %s, got %s", savedSession.RefreshToken, userSessionDto.RefreshToken)
	}
}

func TestGetUserSession_UserNotFound(t *testing.T) {
	service, _, sessionRepo := setupTestService()

	// Create a session with non-existent user ID
	session := &domain.Session{
		UserID: "non-existent-user-id",
	}
	savedSession, err := sessionRepo.Save(session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Try to get user session
	_, err = service.GetUserSession(savedSession)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestRefreshToken_NotImplemented(t *testing.T) {
	service, _, _ := setupTestService()

	_, err := service.RefreshToken("some-refresh-token")
	if err == nil {
		t.Fatal("expected error for not implemented, got nil")
	}

	if err.Error() != "not implemented" {
		t.Errorf("expected 'not implemented' error, got %v", err)
	}
}

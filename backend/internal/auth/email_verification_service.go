package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/google/uuid"
)

const emailVerificationLifetime = 15 * time.Minute

var (
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationExpired     = errors.New("verification code expired")
	ErrEmailAlreadyVerified    = errors.New("email already verified")
)

func generateVerificationCode() (string, error) {
	var randomBytes [4]byte

	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", fmt.Errorf(
			"failed to generate verification code: %w",
			err,
		)
	}

	value := uint32(randomBytes[0])<<24 |
		uint32(randomBytes[1])<<16 |
		uint32(randomBytes[2])<<8 |
		uint32(randomBytes[3])

	code := value % 1000000

	return fmt.Sprintf("%06d", code), nil
}

func hashVerificationCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func (s *Service) createEmailVerification(
	ctx context.Context,
	userID string,
) (string, error) {
	code, err := generateVerificationCode()
	if err != nil {
		return "", err
	}

	verification := &EmailVerification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     hashVerificationCode(code),
		ExpiresAt: time.Now().UTC().Add(emailVerificationLifetime),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.emailVerificationRepository.Create(
		ctx,
		verification,
	); err != nil {
		return "", fmt.Errorf(
			"failed to create email verification: %w",
			err,
		)
	}

	return code, nil
}

func (s *Service) SendVerificationEmail(
	ctx context.Context,
	userID string,
	emailAddress string,
) error {
	code, err := s.createEmailVerification(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to create email verification: %w", err)
	}

	subject := "Verify your STAQ account"

	body := fmt.Sprintf(
		"Hello, \n\n"+
			"Your STAQ email verification code is:\n\n"+
			"%s\n\n"+
			"This code expires in 15 minutes.\n\n"+
			"If you did not create a STAQ account, you can ignore this email.\n\n"+
			"Regards,\n"+
			"STAQ",
		code,
	)

	if err := s.emailSender.Send(
		emailAddress,
		subject,
		body,
	); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *Service) VerifyEmail(
	ctx context.Context,
	emailAddress string,
	code string,
) error {
	if err := validateEmail(emailAddress); err != nil {
		return err
	}

	if len(code) != 6 {
		return ErrInvalidVerificationCode
	}

	user, err := s.usersService.GetByEmail(ctx, emailAddress)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return ErrInvalidVerificationCode
		}

		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	tokenHash := hashVerificationCode(code)

	verification, err := s.emailVerificationRepository.FindByToken(
		ctx,
		tokenHash,
	)
	if err != nil {
		return ErrInvalidVerificationCode
	}

	if verification.UserID != user.ID {
		return ErrInvalidVerificationCode
	}

	if verification.VerifiedAt != nil {
		return ErrInvalidVerificationCode
	}

	if verification.RevokedAt != nil {
		return ErrInvalidVerificationCode
	}

	if !time.Now().UTC().Before(verification.ExpiresAt) {
		return ErrVerificationExpired
	}

	if err := s.emailVerificationRepository.MarkVerified(
		ctx,
		verification.ID,
	); err != nil {
		return fmt.Errorf("failed to verify user email: %w", err)
	}

	if err := s.usersService.MarkEmailVerified(
		ctx,
		user.ID,
	); err != nil {
		return fmt.Errorf("failed to verify user email: %w", err)
	}

	return nil
}

func (s *Service) ResendVerificationEmail(
	ctx context.Context,
	emailAddress string,
) error {
	if err := validateEmail(emailAddress); err != nil {
		return err
	}

	user, err := s.usersService.GetByEmail(ctx, emailAddress)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return ErrInvalidVerificationCode
		}

		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	if err := s.emailVerificationRepository.RevokeForUser(
		ctx,
		user.ID,
	); err != nil {
		return fmt.Errorf("failed to revoke previous verification: %w", err)
	}

	if err := s.SendVerificationEmail(
		ctx,
		user.ID,
		user.Email,
	); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

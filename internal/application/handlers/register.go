package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/domain/aggregates"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"github.com/victorotene80/medilog-api/internal/domain/valueobjects"
)

type RegisterHandler struct {
	userRepo        repository.UserAggregateRepository
	tx              appContracts.TransactionManager
	passwordHasher  contracts.PasswordHasher
	eventPublisher  appContracts.MessagePublisher
	clock           func() time.Time
	passwordService contracts.PasswordService
}

func NewRegisterHandler(
	userRepo repository.UserAggregateRepository,
	tx appContracts.TransactionManager,
	passwordHasher contracts.PasswordHasher,
	eventPublisher appContracts.MessagePublisher,
	clock func() time.Time,
	passwordService contracts.PasswordService,
) *RegisterHandler {
	return &RegisterHandler{
		userRepo:        userRepo,
		tx:              tx,
		passwordHasher:  passwordHasher,
		eventPublisher:  eventPublisher,
		clock:           clock,
		passwordService: passwordService,
	}
}

func (h *RegisterHandler) Handle(
	ctx context.Context,
	cmd command.RegisterCommand,
) (*dto.CreateUserDTO, error) {
	email, phone, err := buildRegisterContact(cmd)
	if err != nil {
		return nil, err
	}

	if err := h.passwordService.Validate(cmd.Password); err != nil {
		return nil, err
	}

	if email != nil {
		exists, err := h.userRepo.ExistsByEmail(ctx, email.String())
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, application.NewConflict("email already registered")
		}
	}
	if phone != nil {
		exists, err := h.userRepo.ExistsByPhone(ctx, phone.String())
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, application.NewConflict("phone already registered")
		}
	}

	countryCode, err := valueobjects.NewCountryCode(cmd.CountryCode)
	if err != nil {
		return nil, application.NewValidation("invalid country code")
	}
	sex, err := valueobjects.NewSex(cmd.Sex)
	if err != nil {
		return nil, application.NewValidation(err.Error())
	}
	bloodType, err := valueobjects.NewBloodType(cmd.BloodType)
	if err != nil {
		return nil, application.NewValidation(err.Error())
	}
	dateOfBirth, err := time.Parse("2006-01-02", strings.TrimSpace(cmd.DOB))
	if err != nil {
		return nil, application.NewValidation("invalid date of birth")
	}
	if err := validateDateOfBirth(dateOfBirth, h.clock()); err != nil {
		return nil, err
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	emailString, phoneString := registerContactToStrings(email, phone)

	user := entities.NewUser(
		emailString,
		phoneString,
		strings.TrimSpace(cmd.FirstName),
		strings.TrimSpace(cmd.LastName),
		countryCode,
		&passwordHash,
		dateOfBirth,
		bloodType,
		sex,
	)

	agg := aggregates.NewUserAggregate(user)

	// The device timezone places reminders at the user's local wall-clock time.
	// An unrecognized name keeps the UTC default rather than failing sign-up at
	// its last step; the user can still correct it from their profile.
	if cmd.Timezone != nil {
		if tz, err := valueobjects.NewTimezone(*cmd.Timezone); err == nil {
			agg.Profile.Timezone = tz
		}
	}

	// The user row and the user.created outbox row commit together or not at
	// all. Previously Save committed on its own connection and the outbox write
	// followed on another with its error discarded, so a crash in between lost
	// the event while the API still returned 201.
	if err := h.tx.Do(ctx, func(ctx context.Context) error {
		if err := h.userRepo.Save(ctx, agg); err != nil {
			return err
		}

		// Raised after Save so the event carries the ID the database assigned.
		agg.RaiseCreatedEvent()
		domainEvents := agg.PullEvents()

		if h.eventPublisher == nil || len(domainEvents) == 0 {
			return nil
		}

		// Without explicit metadata ToEnvelope falls back to the domain event
		// name ("user.created"), which matches no binding — the broker discards
		// the message and the relay still marks the row sent.
		eventMeta := messaging.Context{
			Kind:          messaging.KindIntegrationEvent,
			Name:          messaging.EventUserCreated,
			AggregateType: "user",
			Action:        "register",
		}

		return h.eventPublisher.Publish(ctx, domainEvents, eventMeta.ToMetadata())
	}); err != nil {
		return nil, err
	}

	return &dto.CreateUserDTO{
		UserID:              user.ID,
		Email:               emailString,
		Phone:               phoneString,
		FirstName:           strings.TrimSpace(cmd.FirstName),
		LastName:            strings.TrimSpace(cmd.LastName),
		OnboardingCompleted: false,
		RequiresOnboarding:  true,
	}, nil
}

func buildRegisterContact(
	cmd command.RegisterCommand,
) (*valueobjects.Email, *valueobjects.Phone, error) {
	var email *valueobjects.Email
	var phone *valueobjects.Phone

	if cmd.Email != nil && strings.TrimSpace(*cmd.Email) != "" {
		emailVO, err := valueobjects.NewEmail(*cmd.Email)
		if err != nil {
			return nil, nil, application.NewValidation("invalid email address")
		}
		email = &emailVO
	}

	if cmd.Phone != nil && strings.TrimSpace(*cmd.Phone) != "" {
		phoneVO, err := valueobjects.NewPhone(*cmd.Phone)
		if err != nil {
			return nil, nil, application.NewValidation("invalid phone number")
		}
		phone = &phoneVO
	}

	if email == nil && phone == nil {
		return nil, nil, application.NewValidation("email or phone number is required")
	}

	return email, phone, nil
}

func registerContactToStrings(
	email *valueobjects.Email,
	phone *valueobjects.Phone,
) (*string, *string) {
	var emailString *string
	var phoneString *string

	if email != nil {
		v := email.String()
		emailString = &v
	}
	if phone != nil {
		v := phone.String()
		phoneString = &v
	}

	return emailString, phoneString
}

// minDateOfBirth rejects slips of the picker wheel (1874 was accepted in QA)
// while still admitting the oldest living people.
var minDateOfBirth = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func validateDateOfBirth(dob, now time.Time) error {
	if dob.Before(minDateOfBirth) || dob.After(now) {
		return application.NewValidation("date of birth must be between 1900-01-01 and today")
	}
	return nil
}

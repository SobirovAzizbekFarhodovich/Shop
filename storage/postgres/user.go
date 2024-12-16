package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	pb "auth/genprotos"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

const emailRegex = `^[a-zA-Z0-9._]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
const PhoneNumberRegex = `^\+?998(90|91|93|94|97|88|98|33|71)\d{7}$|^8(90|91|93|94|97|88|98|33|71)\d{7}$`

func isValidEmail(email string) bool {
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidPhoneNumber(phone_number string) bool {
	re := regexp.MustCompile(PhoneNumberRegex)
	return re.MatchString(phone_number)
}

func (u *UserStorage) RegisterUser(user *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	if !isValidEmail(user.Email) {
		return nil, errors.New("invalid email format")
	}

	if !isValidPhoneNumber(user.PhoneNumber) {
		return nil, errors.New("invalid phone number format")
	}

	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := u.db.QueryRow(queryCheck, user.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %v", err)
	}

	if exists {
		return nil, errors.New("user already registered")
	}
	query := `
        INSERT INTO users (email, password, full_name, profile_picture, bio, phone_number, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`
	var userID string
	err = u.db.QueryRow(
		query,
		user.Email,
		string(user.Password),
		user.FullName,
		user.ProfilePicture,
		user.Bio,
		user.PhoneNumber,
		time.Now(),
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}
	response := &pb.RegisterUserResponse{
		Id: userID,
	}
	return response, nil
}

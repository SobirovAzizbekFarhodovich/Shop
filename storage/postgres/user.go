package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
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

func (u *UserStorage) LoginUser(user *pb.LoginUserRequest)(*pb.LoginUserResponse, error){
	query := `SELECT id, email, password, full_name, profile_picture, bio, phone_number, role FROM users WHERE email = $1 and deleted_at = 0`
	row := u.db.QueryRow(query, user.Email)
	res := pb.LoginUserResponse{}
	err := row.Scan(
		&res.Id,
		&res.Email,
		&res.Password,
		&res.FullName,
		&res.ProfilePicture,
		&res.Bio,
		&res.PhoneNumber,
		&res.Role,
	)
	if err != nil{
		if err == sql.ErrNoRows{
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}
	return &res, nil
}

func (u *UserStorage) GetByIdUser(id *pb.GetByIdUserRequest)(*pb.GetByIdUserResponse, error){
	query := `SELECT email, full_name,profile_picture,bio,phone_number FROM users WHERE id = $1 and deleted_at = 0`
	row := u.db.QueryRow(query,id.Id)
	user := pb.GetByIdUserResponse{}
	err := row.Scan(
		&user.Email,
		&user.FullName,
		&user.ProfilePicture,
		&user.Bio,
		&user.PhoneNumber,
	)
	if err != nil{
		if err == sql.ErrNoRows{
			return nil, errors.New("users not found")
		}
		return nil, err
	}
	return &user, err
}

func (u *UserStorage) UpdateUser(user *pb.UpdateUserRequest)(*pb.UpdateUserResponse, error){
	query := `UPDATE users SET `
	var condition []string
	var args []interface{}

	if user.Bio != "" && user.Bio != "string"{
		condition = append(condition, fmt.Sprintf("bio = $%d",len(args) + 1))
		args = append(args, user.Bio)
	}
	if user.Email != "" && user.Email != "string"{
		condition = append(condition, fmt.Sprintf("email = $%d", len(args) + 1))
		args = append(args, user.Email)
	}
	if user.FullName != "" && user.FullName != "string"{
		condition = append(condition, fmt.Sprintf("full_name = $%d", len(args) + 1))
		args = append(args, user.FullName)
	}
	if user.ProfilePicture != "" && user.ProfilePicture != "string"{
		condition = append(condition, fmt.Sprintf("profile_picture = $%d", len(args) + 1))
		args = append(args, user.ProfilePicture)
	}
	if len(condition) == 0{
		return nil, errors.New("nothing to update")
	}

	query += strings.Join(condition, ", ")
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, bio, email, full_name, profile_picture", len(args) + 1)
	args = append(args, user.Id)

	res := pb.UpdateUserResponse{}
	row := u.db.QueryRow(query,args...)
	err := row.Scan(
		&res.Id,
		&res.Bio,
		&res.Email,
		&res.FullName,
		&res.ProfilePicture,
	)
	if err != nil{
		return nil, err
	}
	return &res, nil
}

func (u *UserStorage) DeleteUser(user *pb.DeleteUserRequest)(*pb.DeleteUserResponse, error){
	query := `UPDATE users SET deleted_at = $2 WHERE id = $1 and deleted_at = 0`
	_, err := u.db.Exec(query,user.Id, time.Now().Unix())
	if err != nil{
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}
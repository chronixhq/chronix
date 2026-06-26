package cxuser

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"fmt"
	"strings"

	"github.com/dan-sherwin/go-utilities"
	"golang.org/x/crypto/bcrypt"
)

type (
	CxUser struct {
		models.CxUser
		AuthKey string `json:"-" gorm:"-"`
	}
)

// GetUserByID returns the full user row by ID or nil if not found.
func GetUserByID(id int64) (*CxUser, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid id")
	}
	u, err := db.CxUser.Where(db.CxUser.ID.Eq(id)).Take()
	if err != nil || u == nil {
		return nil, err
	}
	return &CxUser{CxUser: *u}, nil
}

func Login(email, password string) (*CxUser, error) {
	user, err := db.CxUser.Where(
		db.CxUser.Email.Lower().Eq(strings.ToLower(strings.TrimSpace(email))),
		db.CxUser.Enabled.Is(true),
		db.CxUser.Suspended.Is(false),
	).Take()
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.Password == nil {
		return nil, fmt.Errorf("password is required")
	}

	// Try bcrypt first
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err == nil {
		return &CxUser{CxUser: *user}, nil
	}

	return nil, fmt.Errorf("invalid password")
}

func encryptPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func (c *CxUser) IncrementSv() error {
	if c.ID == 0 {
		return nil
	}
	user, err := db.CxUser.Where(db.CxUser.ID.Eq(c.ID)).Take()
	if err != nil {
		return err
	}
	user.Sv++
	return db.CxUser.Save(user)
}

func (c *CxUser) Save() error {
	if c.Email == "" || !utilities.IsEmail(c.Email) {
		return fmt.Errorf("valid email is required")
	}
	if len(c.Name) < 3 {
		return fmt.Errorf("name must be at least 3 characters")
	}
	if c.Password != nil && len(strings.TrimSpace(*c.Password)) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	// Allow creating a user without an initial password; they must set it later
	if c.Password != nil {
		trimmedPass := strings.TrimSpace(*c.Password)
		encrypted := encryptPassword(trimmedPass)
		c.Password = &encrypted
	}
	if c.ID == 0 {
		if _, err := db.CxUser.Where(db.CxUser.Email.Eq(c.Email)).Take(); err == nil {
			return fmt.Errorf("user already exists")
		}
		c.Enabled = true
		err := db.CxUser.Create(&c.CxUser)
		if err != nil {
			return fmt.Errorf("error creating user: %v", err)
		}
		return nil
	}
	existingUser, err := db.CxUser.Where(db.CxUser.ID.Eq(c.ID)).Take()
	if err != nil {
		return fmt.Errorf("error retrieving user: %v", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found")
	}
	// Preserve immutable/managed fields and any unspecified optional fields
	c.ID = existingUser.ID
	if c.Password == nil {
		c.Password = existingUser.Password
	}
	// Preserve optional pointer fields if not provided in update to avoid clearing them
	if c.Phone == nil {
		c.Phone = existingUser.Phone
	}
	if c.TimeZone == nil {
		c.TimeZone = existingUser.TimeZone
	}
	if c.TimeFormat == nil {
		c.TimeFormat = existingUser.TimeFormat
	}
	// Enabled/Admin/Suspended flags are managed separately by admin workflows
	c.Enabled = existingUser.Enabled
	c.Admin = existingUser.Admin
	c.Suspended = existingUser.Suspended
	err = db.CxUser.Save(&c.CxUser)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}
	return nil
}

func (c *CxUser) IsAdmin() bool {
	return c.Admin
}

func (c *CxUser) SetAdmin(admin bool) {
	c.Admin = admin
	_ = db.CxUser.Save(&c.CxUser)
}

func (c *CxUser) SetEnabled(enabled bool) {
	c.Enabled = enabled
	_ = db.CxUser.Save(&c.CxUser)
}

func (c *CxUser) Delete() error {
	if c.ID == 0 {
		return fmt.Errorf("cannot delete user with ID 0")
	}
	_, err := db.CxUser.Where(db.CxUser.ID.Eq(c.ID)).Delete()
	return err
}

func UserList() ([]*CxUser, error) {
	users, err := db.CxUser.Find()
	if err != nil {
		return nil, err
	}
	cxUsers := make([]*CxUser, len(users))
	for i, user := range users {
		user.Password = nil
		cxUsers[i] = &CxUser{CxUser: *user}
	}
	return cxUsers, nil
}

// IsEmailAvailable returns true if no other user has the given email (case-insensitive).
// If excludeID > 0, the user with that ID will be excluded from the uniqueness check
// (useful when updating an existing user's email).
func IsEmailAvailable(email string, excludeID int64) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, nil
	}
	q := db.CxUser.Where(db.CxUser.Email.Lower().Eq(email))
	if excludeID > 0 {
		q = q.Where(db.CxUser.ID.Neq(excludeID))
	}
	existing, err := q.Take()
	if err != nil {
		// If not found, it's available
		return true, nil
	}
	return existing == nil, nil
}

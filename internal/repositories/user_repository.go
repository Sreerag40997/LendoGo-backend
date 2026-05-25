package repositories

import (
	// "lendogo-backend/internal/structures/models"
	// "lendogo-backend/internal/models"

	"lendogo-backend/structures/models"

	"gorm.io/gorm"
)

// 1. The Interface (The ADT definition)
type UserRepository interface {
	CreateUser(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	UpdatePassword(email string, hashedPassword string) error

}

// 2. The Struct (Holds the DB connection)
type userRepositoryImpl struct {
	db *gorm.DB
}


// 3. The Constructor (Used for Dependency Injection)
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// 4. The Implementations
func (r *userRepositoryImpl) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user models.User
	// Fetch the first record that matches the email
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepositoryImpl) UpdatePassword(email string, hashedPassword string) error {
    return r.db.Model(&models.User{}).
        Where("email = ?", email).
        Update("password", hashedPassword).Error
}

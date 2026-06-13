package repositories

import (
    "errors"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "lendogo-backend/structures/models"
)

type AdminRepository interface {
    // --- USER MANAGEMENT (External Borrowers) ---
    GetAllUsers() ([]models.User, error)
    CreateSystemUserTx(user *models.User, profile *models.UserProfile) error
    UpdateUser(id string, updates map[string]interface{}) error
    HardDeleteUser(userID uuid.UUID) error
    UpdateUserStatus(id string, status string) error

    // --- STAFF MANAGEMENT (Internal Employees) ---
    CreateStaff(staff *models.Staff) error
    GetAllStaff() ([]models.Staff, error)
    GetStaffByEmail(email string) (*models.Staff, error) // 👈 ADDED for Login functionality
    HardDeleteStaff(staffID uuid.UUID) error
    UpdateStaffStatus(id string, status string) error
}

type adminRepoImpl struct {
    db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
    return &adminRepoImpl{db: db}
}

// ==========================================
// 1. USER MANAGEMENT (External Borrowers)
// ==========================================

func (r *adminRepoImpl) GetAllUsers() ([]models.User, error) {
    var users []models.User
    err := r.db.Select("id, full_name, email, role, status, created_at").Order("created_at DESC").Find(&users).Error
    return users, err
}

func (r *adminRepoImpl) CreateSystemUserTx(user *models.User, profile *models.UserProfile) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        return nil
    })
}

func (r *adminRepoImpl) UpdateUser(id string, updates map[string]interface{}) error {
    return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *adminRepoImpl) HardDeleteUser(userID uuid.UUID) error {
    result := r.db.Unscoped().Where("id = ?", userID).Delete(&models.User{})
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }
    return nil
}

func (r *adminRepoImpl) UpdateUserStatus(id string, status string) error {
    return r.db.Exec("UPDATE users SET status = ? WHERE id = ?", status, id).Error
}

// ==========================================
// 2. STAFF MANAGEMENT (Internal Employees)
// ==========================================

func (r *adminRepoImpl) CreateStaff(staff *models.Staff) error {
    return r.db.Create(staff).Error
}

func (r *adminRepoImpl) GetAllStaff() ([]models.Staff, error) {
    var staffList []models.Staff
    err := r.db.Order("created_at DESC").Find(&staffList).Error
    return staffList, err
}

// GetStaffByEmail fetches a staff record by email for authentication
func (r *adminRepoImpl) GetStaffByEmail(email string) (*models.Staff, error) {
    var staff models.Staff
    err := r.db.Where("email = ?", email).First(&staff).Error
    if err != nil {
        return nil, err
    }
    return &staff, nil
}

func (r *adminRepoImpl) HardDeleteStaff(staffID uuid.UUID) error {
    result := r.db.Unscoped().Where("id = ?", staffID).Delete(&models.Staff{})
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("staff not found")
    }
    return nil
}

func (r *adminRepoImpl) UpdateStaffStatus(id string, status string) error {
    return r.db.Exec("UPDATE staffs SET status = ? WHERE id = ?", status, id).Error
}
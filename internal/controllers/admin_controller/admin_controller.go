package admin_controller

import (
    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    "lendogo-backend/database"
    "lendogo-backend/structures/models"
    "lendogo-backend/utils" // 👈 ADDED: Import your utils package for AWS Presigner
)

// AdminController structure handles administrative HTTP requests.
type AdminController struct {
    // TODO: Transition from global database state (database.DB) to dependency injected repositories.
}

// NewAdminController initializes a new AdminController.
func NewAdminController() *AdminController {
    return &AdminController{}
}

// GetAllUsers retrieves all system users (Admin restricted).
func (c *AdminController) GetAllUsers(ctx *fiber.Ctx) error {
    var users []models.User

    // Fetch users from the database, ordering by the newest first.
    // We omit passwords for security!
    result := database.DB.Select("id, full_name, email, role, created_at").
        Order("created_at DESC").
        Find(&users)

    if result.Error != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch users from database",
        })
    }

    return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Users fetched successfully",
        "data":    users,
    })
}

// GetSystemStats retrieves aggregated system metrics.
func (c *AdminController) GetSystemStats(ctx *fiber.Ctx) error {
    // TODO: Replace mock with actual DB aggregations
    return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
        "message":      "System is running perfectly.",
        "active_loans": 42,
    })
}

// GetAllApplications fetches all loan records, preloading 1:1 KYC and Financial associations.
func (c *AdminController) GetAllApplications(ctx *fiber.Ctx) error {
    var applications []models.LoanApplication

    // Preload executes LEFT JOINs (or concurrent IN queries) to resolve nested associations efficiently, preventing N+1 issues.
    result := database.DB.
        Preload("KYC").
        Preload("FinancialDocs").
        Order("created_at DESC").
        Find(&applications)

    if result.Error != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to execute data aggregation for applications",
        })
    }

    // 👇 ADDED: The AWS Cryptographic Signing Loop 👇
    // Before sending the data to React, we convert the dead, private S3 URLs 
    // into secure, 15-minute temporary viewing links.
    for i := range applications {
        // Sign KYC Documents
        applications[i].KYC.LiveSelfiePath = utils.GeneratePresignedURL(applications[i].KYC.LiveSelfiePath)
        applications[i].KYC.AadhaarFrontPath = utils.GeneratePresignedURL(applications[i].KYC.AadhaarFrontPath)
        applications[i].KYC.AadhaarBackPath = utils.GeneratePresignedURL(applications[i].KYC.AadhaarBackPath)
        applications[i].KYC.PanCardPath = utils.GeneratePresignedURL(applications[i].KYC.PanCardPath)

        // Sign Financial Documents
        applications[i].FinancialDocs.BankStatementPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.BankStatementPath)
        applications[i].FinancialDocs.PropertyAgreemntPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.PropertyAgreemntPath)
        applications[i].FinancialDocs.IncomeProofPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.IncomeProofPath)
    }
    // 👆 END AWS LOOP 👆

    return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Applications fetched successfully",
        "data":    applications,
    })
}

// UpdateApplicationStatus mutates the status of a specific loan application via a state machine.
func (c *AdminController) UpdateApplicationStatus(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    
    var payload struct {
        Status string `json:"status"`
    }

    if err := ctx.BodyParser(&payload); err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed JSON payload"})
    }

    // Enforce strict state transitions. Do not allow arbitrary strings.
    validStates := map[string]bool{
        "APPROVED":                 true, 
        "REJECTED":                 true, 
        "ADDITIONAL_DOCS_REQUIRED": true,
        "DISBURSED":                true,
    }
    
    if !validStates[payload.Status] {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid state transition requested"})
    }

    // If the status is DISBURSED, perform atomic double-entry balance adjustment
    if payload.Status == "DISBURSED" {
        err := database.DB.Transaction(func(tx *gorm.DB) error {
            var loan models.LoanApplication
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&loan).Error; err != nil {
                return err
            }

            // If it is already disbursed, do nothing
            if loan.ApplicationStatus == "DISBURSED" {
                return nil
            }

            // Lock System Wallet
            var sysWallet models.SystemWallet
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_name = ?", "capital_disbursement").First(&sysWallet).Error; err != nil {
                return err
            }

            if sysWallet.Balance < loan.PrincipalAmount {
                return fiber.NewError(fiber.StatusBadRequest, "Insufficient capital reserves in system wallet")
            }

            // Debit system wallet
            if err := tx.Model(&sysWallet).UpdateColumn("balance", gorm.Expr("balance - ?", loan.PrincipalAmount)).Error; err != nil {
                return err
            }

            // Credit borrower wallet
            var userWallet models.UserWallet
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", loan.UserID).FirstOrCreate(&userWallet, models.UserWallet{UserID: loan.UserID, Balance: 0}).Error; err != nil {
                return err
            }

            if err := tx.Model(&userWallet).UpdateColumn("balance", gorm.Expr("balance + ?", loan.PrincipalAmount)).Error; err != nil {
                return err
            }

            // Create ledger entries
            entries := []models.LedgerEntry{
                {
                    WalletID:        sysWallet.ID,
                    Amount:          -loan.PrincipalAmount,
                    TransactionType: "DISBURSEMENT_DEBIT",
					ReferenceID:     loan.ID.String(),
                },
                {
                    WalletID:        userWallet.ID,
					Amount:          loan.PrincipalAmount,
                    TransactionType: "LOAN_CREDIT",
					ReferenceID:     loan.ID.String(),
                },
            }
            if err := tx.Create(&entries).Error; err != nil {
                return err
            }

            // Set status to DISBURSED
            return tx.Model(&loan).UpdateColumn("application_status", "DISBURSED").Error
        })

        if err != nil {
            if fiberErr, ok := err.(*fiber.Error); ok {
                return ctx.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
            }
            return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }

        return ctx.SendStatus(fiber.StatusOK)
    }

    // For other statuses, simply execute the partial update on the application_status column
    result := database.DB.Model(&models.LoanApplication{}).
        Where("id = ?", id).
        Update("application_status", payload.Status)

    if result.Error != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit state change"})
    }

    if result.RowsAffected == 0 {
        return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "loan application reference not found"})
    }

    return ctx.SendStatus(fiber.StatusOK)
}

// GetAllConsultations retrieves all consultation requests (Admin restricted).
func (c *AdminController) GetAllConsultations(ctx *fiber.Ctx) error {
    var consultations []models.Consultation

    result := database.DB.Order("created_at DESC").Find(&consultations)

    if result.Error != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch consultations from database",
        })
    }

    return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Consultations fetched successfully",
        "data":    consultations,
    })
}
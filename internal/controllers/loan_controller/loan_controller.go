package loan_controller

import (
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"lendogo-backend/internal/services"
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"
)

type LoanController struct {
	service services.LoanService
}

func NewLoanController(service services.LoanService) *LoanController {
	return &LoanController{service: service}
}

// ApplyForLoan handles the multipart/form-data request from React
func (c *LoanController) ApplyForLoan(ctx *fiber.Ctx) error {
	// 1. Get the User ID from the secure HttpOnly cookie
	userID := ctx.Locals("user_id").(string)

	// 2. Parse the text fields from the form
	principal, _ := strconv.ParseFloat(ctx.FormValue("principal_amount"), 64)
	tenure, _ := strconv.Atoi(ctx.FormValue("tenure_months"))
	income, _ := strconv.ParseFloat(ctx.FormValue("monthly_income"), 64)

	// 3. Build the base Loan Application model
	loanApp := &models.LoanApplication{
		UserID:           uuid.MustParse(userID),
		FullName:         ctx.FormValue("full_name"),
		DOB:              ctx.FormValue("dob"),
		Email:            ctx.FormValue("email"),
		MobileNumber:     ctx.FormValue("mobile_number"),
		Address:          ctx.FormValue("address"),
		City:             ctx.FormValue("city"),
		State:            ctx.FormValue("state"),
		Pincode:          ctx.FormValue("pincode"),
		LoanTrack:        ctx.FormValue("loan_track"),
		ProductCategory:  ctx.FormValue("product_category"),
		PrincipalAmount:  principal,
		TenureMonths:     tenure,
	}
	loanApp.FinancialDocs.EmploymentStatus = ctx.FormValue("employment_status")
	loanApp.FinancialDocs.MonthlyIncome = income


	// 4. Handle File Uploads securely via S3
	// We check if the file exists in the request. If it does, we upload it!
	if file, err := ctx.FormFile("live_selfie"); err == nil {
		if url, err := utils.UploadFileToS3(file); err == nil {
			loanApp.KYC.LiveSelfiePath = url
		}
	}
	if file, err := ctx.FormFile("aadhaar_front"); err == nil {
		if url, err := utils.UploadFileToS3(file); err == nil {
			loanApp.KYC.AadhaarFrontPath = url
		}
	}
	if file, err := ctx.FormFile("aadhaar_back"); err == nil {
		if url, err := utils.UploadFileToS3(file); err == nil {
			loanApp.KYC.AadhaarBackPath = url
		}
	}
	if file, err := ctx.FormFile("pan_card"); err == nil {
		if url, err := utils.UploadFileToS3(file); err == nil {
			loanApp.KYC.PanCardPath = url
		}
	}
	if file, err := ctx.FormFile("bank_statement"); err == nil {
		if url, err := utils.UploadFileToS3(file); err == nil {
			loanApp.FinancialDocs.BankStatementPath = url
		}
	}

	// 5. Send it to the Service to do the math and save to the Database!
	err := c.service.ProcessApplication(loanApp)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process loan application",
		})
	}

	// 6. SUCCESS! Send the Reference Number back to React
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":          "Loan Application Submitted Successfully!",
		"reference_number": loanApp.ReferenceNumber,
		"estimated_emi":    loanApp.EstimatedEMI,
		"status":           loanApp.ApplicationStatus,
	})
}
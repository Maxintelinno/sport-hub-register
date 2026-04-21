package repository

import ( 
	"math"
	"sport-hub-register/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreditRepository struct {
	db *gorm.DB
}

func NewCreditRepository(db *gorm.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) GetByUserID(tx *gorm.DB, userID string) (*model.BookingCredit, error) {
	var credit model.BookingCredit
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.Where("user_id = ?", userID).First(&credit).Error
	if err != nil {
		return nil, err
	}
	return &credit, nil
}

func (r *CreditRepository) AdjustBalance(tx *gorm.DB, userID uuid.UUID, amount float64, transType string, sourceType string, referenceID *uuid.UUID, note string) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	var credit model.BookingCredit
	if err := db.Where(model.BookingCredit{UserID: userID}).FirstOrCreate(&credit).Error; err != nil {
		return err
	}

	balanceBefore := credit.Balance
	credit.Balance += amount

	if amount > 0 {
		credit.TotalEarned += amount
	} else if amount < 0 {
		credit.TotalUsed += math.Abs(amount)
	}

	credit.UpdatedAt = time.Now()
	if err := db.Save(&credit).Error; err != nil {
		return err
	}

	transaction := model.BookingCreditTransaction{
		ID:              uuid.New(),
		UserID:          userID,
		BookingCreditID: credit.ID,
		TransactionType: transType,
		SourceType:      sourceType,
		ReferenceID:     referenceID,
		Amount:          math.Abs(amount),
		BalanceBefore:   balanceBefore,
		BalanceAfter:    credit.Balance,
		Note:            note,
		CreatedAt:       time.Now(),
	}

	return db.Create(&transaction).Error
}

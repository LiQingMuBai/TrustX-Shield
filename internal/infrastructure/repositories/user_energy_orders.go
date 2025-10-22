package repositories

import (
	"context"
	_ "github.com/go-sql-driver/mysql"

	"gorm.io/gorm"
	"ushield_bot/internal/domain"
)

type UserEnergyOrdersRepo struct {
	db *gorm.DB
}

func NewUserEnergyOrdersRepo(db *gorm.DB) *UserEnergyOrdersRepo {
	return &UserEnergyOrdersRepo{
		db: db,
	}
}

func (r *UserEnergyOrdersRepo) Create(ctx context.Context, userAddress *domain.UserEnergyOrders) error {
	return r.db.WithContext(ctx).Create(userAddress).Error
}

func (r *UserEnergyOrdersRepo) Count(ctx context.Context, _chatID int64) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserEnergyOrders{}).Where("chat_id = ?", _chatID).Count(&count).Error
	if err != nil {
		return
	}
	return count, nil
}

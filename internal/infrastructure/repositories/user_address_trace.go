package repositories

import (
	"context"

	_ "github.com/go-sql-driver/mysql"

	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserAddressTraceRepo struct {
	db *gorm.DB
}

func NewUserAddressTraceRepo(db *gorm.DB) *UserAddressTraceRepo {
	return &UserAddressTraceRepo{
		db: db,
	}
}

func (r *UserAddressTraceRepo) Create(ctx context.Context, userAddress *domain.UserAddressTrace) error {
	return r.db.WithContext(ctx).Create(userAddress).Error
}

func (r *UserAddressTraceRepo) Remove(ctx context.Context, _chatID int64, _address string) error {
	//return r.db.WithContext(ctx).del(userAddress).Error

	return r.db.WithContext(ctx).Delete(&domain.UserAddressTrace{}, "chat_id = ? AND address = ?", _chatID, _address).Error
}

//	func (r *UserAddressTraceRepo) Find(ctx context.Context, _chatID int64, _address string) error {
//		//return r.db.WithContext(ctx).del(userAddress).Error
//
//		return r.db.WithContext(ctx).Find(&domain.UserAddressTrace{}, "chat_id = ? AND address = ?", _chatID, _address).Error
//	}
func (r *UserAddressTraceRepo) Find(ctx context.Context, _chatID int64, _address string) (domain.UserAddressTrace, error) {
	var item domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Find(&item, "chat_id = ? AND address = ?", _chatID, _address).Error
	return item, err

	//err := r.db.WithContext(ctx).
	//	Find(&placeholders, "status = ?", 0).Error
	//return placeholders, err

}

//func (r *UserAddressTraceRepo) Count(ctx context.Context, _chatID int64, _address string) int64 {
//	//return r.db.WithContext(ctx).del(userAddress).Error
//
//	return r.db.WithContext(ctx).Count(&domain.UserAddressTrace{}, "chat_id = ? ", _chatID, _address)
//}
//

func (r *UserAddressTraceRepo) Count(ctx context.Context, _chatID int64) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserAddressTrace{}).Where("chat_id = ?", _chatID).Count(&count).Error
	if err != nil {
		return
	}
	return count, nil
}

func (r *UserAddressTraceRepo) Query(ctx context.Context, _chatID int64) ([]domain.UserAddressTrace, error) {
	var subscriptions []domain.UserAddressTrace
	err := r.db.WithContext(ctx).
		Model(&domain.UserAddressTrace{}).
		Select("id", "address", "network").
		Where("chat_id = ?", _chatID).
		Scan(&subscriptions).Error
	return subscriptions, err

}

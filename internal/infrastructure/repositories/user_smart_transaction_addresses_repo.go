package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserSmartTransactionAddressesRepository struct {
	db *gorm.DB
}

func NewUserSmartTransactionAddressesRepository(db *gorm.DB) *UserSmartTransactionAddressesRepository {
	return &UserSmartTransactionAddressesRepository{
		db: db,
	}
}

func (r *UserSmartTransactionAddressesRepository) Create(ctx context.Context, userAddress *domain.UserSmartTransactionAddresses) error {
	return r.db.WithContext(ctx).Create(userAddress).Error
}
func (r *UserSmartTransactionAddressesRepository) Delete(ctx context.Context, _chatID string, _address string) error {
	return r.db.WithContext(ctx).Delete(&domain.UserSmartTransactionAddresses{}, "chat_id = ? AND address = ?", _chatID, _address).Error
}

func (r *UserSmartTransactionAddressesRepository) Remove(ctx context.Context, chatID string, address string) error {
	r.db.WithContext(ctx).Model(&domain.UserSmartTransactionAddresses{}).
		Where("chat_id = ?", chatID).Where("address = ?", address).
		Update("status", 4)
	return nil

}

func (r *UserSmartTransactionAddressesRepository) Enable(ctx context.Context, chatID string, address string) error {
	r.db.WithContext(ctx).Model(&domain.UserSmartTransactionAddresses{}).
		Where("chat_id = ?", chatID).Where("address = ?", address).Where("status = ?", 3).
		Update("status", 1)
	return nil

}

func (r *UserSmartTransactionAddressesRepository) Enable2(ctx context.Context, chatID string, address string) error {
	r.db.WithContext(ctx).Model(&domain.UserSmartTransactionAddresses{}).
		Where("chat_id = ?", chatID).Where("address = ?", address).Where("status = ?", 0).
		Update("status", 1)
	return nil

}

func (r *UserSmartTransactionAddressesRepository) Disable(ctx context.Context, chatID string, address string) error {
	r.db.WithContext(ctx).Model(&domain.UserSmartTransactionAddresses{}).
		Where("chat_id = ?", chatID).Where("address = ?", address).Where("status = ?", 1).
		Update("status", 3)
	return nil

}
func (r *UserSmartTransactionAddressesRepository) ListByToken(ctx context.Context, _token string) ([]domain.UserSmartTransactionAddresses, error) {
	var bundles []domain.UserSmartTransactionAddresses
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionAddresses{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).Where("token = ?", _token).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionAddressesRepository) List(ctx context.Context, _chatID string) ([]domain.UserSmartTransactionAddresses, error) {
	var bundles []domain.UserSmartTransactionAddresses
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionAddresses{}).
		Where("status != 4 and chat_id = ?", _chatID).Order("id desc").
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionAddressesRepository) Find(ctx context.Context, _id string) (domain.UserSmartTransactionAddresses, error) {
	var record domain.UserSmartTransactionAddresses
	err := r.db.WithContext(ctx).
		Find(&record, "id = ?", _id).Error
	return record, err

}
func (r *UserSmartTransactionAddressesRepository) Query(ctx context.Context, _address string) (domain.UserSmartTransactionAddresses, error) {
	var record domain.UserSmartTransactionAddresses
	err := r.db.WithContext(ctx).
		Find(&record, "status != 4 and address = ?", _address).Error
	return record, err
}

func (r *UserSmartTransactionAddressesRepository) Count(ctx context.Context, _chatID int64) (count int64, err error) {
	err = r.db.WithContext(ctx).Model(&domain.UserSmartTransactionAddresses{}).Where("status != 4 and chat_id = ?", _chatID).Count(&count).Error
	if err != nil {
		return
	}
	return count, nil
}

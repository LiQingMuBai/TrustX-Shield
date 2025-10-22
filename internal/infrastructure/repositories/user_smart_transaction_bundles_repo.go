package repositories

import (
	"context"
	"ushield_bot/internal/domain"

	"gorm.io/gorm"
)

type UserSmartTransactionBundlesRepository struct {
	db *gorm.DB
}

func NewUserSmartTransactionBundlesRepository(db *gorm.DB) *UserSmartTransactionBundlesRepository {
	return &UserSmartTransactionBundlesRepository{
		db: db,
	}
}
func (r *UserSmartTransactionBundlesRepository) ListByToken(ctx context.Context, _token string) ([]domain.UserSmartTransactionBundles, error) {
	var bundles []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).Where("token = ?", _token).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionBundlesRepository) ListAll(ctx context.Context) ([]domain.UserSmartTransactionBundles, error) {
	var bundles []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&bundles).Error
	return bundles, err

}
func (r *UserSmartTransactionBundlesRepository) Find(ctx context.Context, _amount string) (domain.UserSmartTransactionBundles, error) {
	var placeholders []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		Select("id", "name").
		Where("amount = ?", _amount).
		Scan(&placeholders).Error
	return placeholders[0], err

}

func (r *UserSmartTransactionBundlesRepository) Query(ctx context.Context, ID string) (domain.UserSmartTransactionBundles, error) {
	var subscriptions []domain.UserSmartTransactionBundles
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionBundles{}).
		//Select("id", "days", "address", "network").
		Where("id = ?", ID).
		Scan(&subscriptions).Error
	return subscriptions[0], err

}

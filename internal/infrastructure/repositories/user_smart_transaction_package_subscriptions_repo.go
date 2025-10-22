package repositories

import (
	"context"
	"errors"
	"fmt"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/request"

	"gorm.io/gorm"
)

type UserSmartTransactionPackageSubscriptionsRepository struct {
	db *gorm.DB
}

func NewUserSmartTransactionPackageSubscriptionsRepository(db *gorm.DB) *UserSmartTransactionPackageSubscriptionsRepository {
	return &UserSmartTransactionPackageSubscriptionsRepository{
		db: db,
	}
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) ListAll(ctx context.Context) ([]domain.UserSmartTransactionPackageSubscriptions, error) {
	var pkgs []domain.UserSmartTransactionPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Select("id", "name", "amount").
		Where("status = ?", 0).
		Scan(&pkgs).Error
	return pkgs, err

}
func (r *UserSmartTransactionPackageSubscriptionsRepository) Query(ctx context.Context, ID string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	var subscriptions []domain.UserSmartTransactionPackageSubscriptions
	err := r.db.WithContext(ctx).
		Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Select("id", "times", "bundle_name", "bundle_id", "amount", "address").
		Where("id = ?", ID).
		Scan(&subscriptions).Error
	return subscriptions[0], err

}

func (r *UserSmartTransactionPackageSubscriptionsRepository) Get(_address string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	record := domain.UserSmartTransactionPackageSubscriptions{}

	err := r.db.Where(" address =? and status = 2", _address).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 记录未找到，不是错误，只是表示不存在
		return record, nil // 第二个返回值表示是否存在
	}

	return record, err
}

func (r *UserSmartTransactionPackageSubscriptionsRepository) GetRecordByID(id string) (domain.UserSmartTransactionPackageSubscriptions, error) {
	record := domain.UserSmartTransactionPackageSubscriptions{}

	err := r.db.Where(" id =? ", id).First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 记录未找到，不是错误，只是表示不存在
		return record, nil // 第二个返回值表示是否存在
	}

	return record, err
}

// Create 创建新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Create(ctx context.Context, pkg *domain.UserSmartTransactionPackageSubscriptions) error {
	return r.db.WithContext(ctx).Create(pkg).Error
}

// Update 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Update(ctx context.Context, pkg *domain.UserSmartTransactionPackageSubscriptions) error {
	return r.db.WithContext(ctx).Save(pkg).Error
}

// Update 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateStatus(ctx context.Context, ID int64, _status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", ID).
		Update("status", _status).Error
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateStatusByID(ctx context.Context, ID string, _status int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", ID).
		Update("status", _status).Error
}

// Update 更新套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) UpdateTimes(ctx context.Context, ID int64, _times int64) error {
	return r.db.WithContext(ctx).Model(&domain.UserSmartTransactionPackageSubscriptions{}).
		Where("id = ?", ID).
		Update("times", _times).Error
}

// Delete 删除套餐
func (r *UserSmartTransactionPackageSubscriptionsRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.UserSmartTransactionPackageSubscriptions{}, id).Error
}
func (r *UserSmartTransactionPackageSubscriptionsRepository) GetUserSmartTransactionPackageSubscriptionsInfoList(ctx context.Context, info request.UserAddressDetectionSearch, _chatID int64) (list []domain.UserSmartTransactionPackageSubscriptions, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	fmt.Printf("limit: %d, offset: %d\n", limit, offset)

	fmt.Printf("page: %d\n", info.Page)
	// 创建db
	db := r.db.Model(&domain.UserSmartTransactionPackageSubscriptions{}).Select("id,status,amount,times,bundle_name,bundle_id,address, DATE_FORMAT(created_at, '%m-%d') as created_date").Where("chat_id = ? and times > 0 and status > 0", _chatID)
	var UserSmartTransactionPackageSubscriptions []domain.UserSmartTransactionPackageSubscriptions
	// 如果有条件搜索 下方会自动创建搜索语句

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(int(limit)).Offset(int(offset)).Order("id DESC")
	}

	err = db.Find(&UserSmartTransactionPackageSubscriptions).Error
	return UserSmartTransactionPackageSubscriptions, total, err
}

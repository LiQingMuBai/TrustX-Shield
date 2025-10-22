package domain

// userSmartTransactionAddresses表 结构体  UserSmartTransactionAddresses
type UserSmartTransactionAddresses struct {
	GVA_MODEL
	Address                  string `json:"address" gorm:"column:address;size:255" form:"address"`
	PlanID                   int    `json:"plan_id" gorm:"column:plan_id" form:"plan_id"`
	PlanNameEn               string `json:"plan_name_en" gorm:"column:plan_name_en;size:255" form:"plan_name_en"`
	PlanNameZh               string `json:"plan_name_zh" gorm:"column:plan_name_zh;size:255" form:"plan_name_zh"`
	PlanSubscriptionSun      int    `json:"plan_subscription_sun" gorm:"column:plan_subscription_sun" form:"plan_subscription_sun"`
	PlanUnitSun              int    `json:"plan_unit_sun" gorm:"column:plan_unit_sun" form:"plan_unit_sun"`
	QuotaMode                string `json:"quota_mode" gorm:"column:quota_mode;size:100" form:"quota_mode"`
	SlotNextBillingTimestamp int    `json:"slot_next_billing_timestamp" gorm:"column:slot_next_billing_timestamp" form:"slot_next_billing_timestamp"`
	Quantity                 int    `json:"quantity" gorm:"column:quantity" form:"quantity"`
	UsedQuantity             int    `json:"used_quantity" gorm:"column:used_quantity" form:"used_quantity"`
	QuotaCount               int    `json:"quota_count" gorm:"column:quota_count" form:"quota_count"`
	UsedCount                int    `json:"used_count" gorm:"column:used_count" form:"used_count"`
	QuotaStartTime           int    `json:"quota_start_time" gorm:"column:quota_start_time" form:"quota_start_time"`
	TotalTransferCount       int    `json:"total_transfer_count" gorm:"column:total_transfer_count" form:"total_transfer_count"`
	TotalEnergyCount         int    `json:"total_energy_count" gorm:"column:total_energy_count" form:"total_energy_count"`
	TotalSlotFee             int    `json:"total_slot_fee" gorm:"column:total_slot_fee" form:"total_slot_fee"`
	TotalEnergyFee           int    `json:"total_energy_fee" gorm:"column:total_energy_fee" form:"total_energy_fee"`
	ChatID                   string `json:"chat_id" gorm:"column:chat_id;size:500" form:"chat_id"`
	IsAutoClosable           bool   `json:"is_auto_closable" gorm:"column:is_auto_closable" form:"is_auto_closable"`
	IdleHours                int    `json:"idle_hours" gorm:"column:idle_hours" form:"idle_hours"`
	Status                   string `json:"status" gorm:"column:status;size:50" form:"status"`
	ExpiredAt                int    `json:"expired_at" gorm:"column:expired_at" form:"expired_at"`
	CreatedDate              int    `json:"created_date" gorm:"column:created_date" form:"created_date"`
}

// TableName userSmartTransactionAddresses表 UserSmartTransactionAddresses自定义表名 user_smart_transaction_addresses
func (UserSmartTransactionAddresses) TableName() string {
	return "user_smart_transaction_addresses"
}

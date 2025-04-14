package models

import "time"

type ReferralBonus struct {
	ID         uint      `gorm:"primaryKey"`
	ReferrerID uint      // The user who referred the new user
	ReferredID uint      // The user who made the deposit (the referred user)
	Amount     float64   // Bonus amount awarded
	RewardedAt time.Time // When the bonus was awarded
}

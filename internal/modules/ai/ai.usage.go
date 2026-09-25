package ai

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultDailyTokens = 200000
	charsPerToken      = 4
)

var (
	ErrQuotaExceeded = errors.New("daily ai limit reached, try again tomorrow")

	usageZone = time.FixedZone("WIB", 7*60*60)
)

type usageStore struct {
	db *gorm.DB
}

func usageDay(now time.Time) (day string, resetsIn time.Duration) {
	local := now.In(usageZone)
	next := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, usageZone)
	return local.Format("2006-01-02"), next.Sub(local)
}

func (u *usageStore) used(userID uuid.UUID, day string) (int64, error) {
	var tokens int64
	err := u.db.Table("ai_usage").
		Select("COALESCE(SUM(tokens), 0)").
		Where("user_id = ? AND day = ?", userID, day).
		Scan(&tokens).Error
	return tokens, err
}

func (u *usageStore) add(userID uuid.UUID, day string, tokens int64) error {
	return u.db.Exec(
		`INSERT INTO ai_usage (user_id, day, tokens) VALUES (?, ?, ?)
		 ON CONFLICT (user_id, day)
		 DO UPDATE SET tokens = ai_usage.tokens + EXCLUDED.tokens, updated_at = now()`,
		userID, day, tokens,
	).Error
}

func estimateTokens(messages []message, output string) int64 {
	chars := utf8.RuneCountInString(output)
	for _, m := range messages {
		chars += utf8.RuneCountInString(m.Content)
	}
	return int64((chars + charsPerToken - 1) / charsPerToken)
}

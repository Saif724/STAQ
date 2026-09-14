package triggers

import "time"

const (
	TypeOnce    = "ONCE"
	TypeDaily   = "DAILY"
	TypeWeekly  = "WEEKLY"
	TypeMonthly = "MONTHLY"
	TypeYearly  = "YEARLY"
	TypeCron    = "CRON"
)

type Trigger struct {
	ID             string     `json:"id"`
	TaskID         string     `json:"task_id"`
	TriggerType    string     `json:"trigger_type"`
	CronExpression *string    `json:"cron_expression,omitempty"`
	TimeZone       string     `json:"timezone"`
	NextRunAt      time.Time  `json:"next_run_at"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

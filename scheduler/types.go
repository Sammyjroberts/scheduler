package scheduler

import (
	"time"
)

type Task struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Priority  float64   `json:"priority"`
}
type ScheduleOutput struct {
	ChosenTasks   []TaskOutput `json:"chosen_tasks"`
	RejectedTasks []TaskOutput `json:"rejected_tasks"`
	TotalPriority float64      `json:"total_priority"`
	Statistics    Statistics   `json:"statistics"`
	TimeRange     TimeRange    `json:"time_range"`
}

type TaskOutput struct {
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Priority       float64 `json:"priority"`
	DurationMins   int     `json:"duration_mins"`
	IsZeroDuration bool    `json:"is_zero_duration"`
}

type Statistics struct {
	TotalTasks     int `json:"total_tasks"`
	ScheduledTasks int `json:"scheduled_tasks"`
	RejectedTasks  int `json:"rejected_tasks"`
}

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

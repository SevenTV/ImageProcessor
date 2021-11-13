package task

import "time"

type TaskEvent struct {
	JobID     string
	Type      TaskEventType
	Timestamp time.Time
}

type TaskEventType string

const (
	Started            TaskEventType = "started"
	Downloaed          TaskEventType = "downloaded"
	Failed             TaskEventType = "failed"
	Completed          TaskEventType = "completed"
	Stopped            TaskEventType = "stopped"
	Cleaned            TaskEventType = "cleaned"
	StageOne           TaskEventType = "stage-one"
	StageOneComplete   TaskEventType = "stage-one-complete"
	StageTwo           TaskEventType = "stage-two"
	StageTwoComplete   TaskEventType = "stage-two-complete"
	StageThree         TaskEventType = "stage-three"
	StageThreeComplete TaskEventType = "stage-three-complete"
)

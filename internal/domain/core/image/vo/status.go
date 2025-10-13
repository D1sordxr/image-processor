package vo

type Status uint // "uploaded", "processing", "completed", "failed"

const (
	StatusUploaded Status = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusUnknown
)

func NewStatus(s string) Status {
	switch s {
	case "uploaded":
		return StatusUploaded
	case "processing":
		return StatusProcessing
	case "completed":
		return StatusCompleted
	case "failed":
		return StatusFailed
	default:
		return StatusUnknown
	}
}

func (s Status) String() string {
	switch s {
	case StatusUploaded:
		return "uploaded"
	case StatusProcessing:
		return "processing"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

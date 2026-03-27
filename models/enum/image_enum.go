package enum

type ImageProvider string

const (
	ImageProviderQiNiu ImageProvider = "qiniu"
)

type ImageUploadTaskStatus string

const (
	ImageUploadTaskPending ImageUploadTaskStatus = "pending"
	ImageUploadTaskReady   ImageUploadTaskStatus = "ready"
	ImageUploadTaskFailed  ImageUploadTaskStatus = "failed"
)

type ImageStatus uint

const (
	ImageStatusReady ImageStatus = iota + 1
	ImageStatusDeleted
	ImageStatusOrphaned
	ImageStatusBlocked
)

func (s ImageStatus) String() string {
	switch s {
	case ImageStatusReady:
		return "ready"
	case ImageStatusDeleted:
		return "deleted"
	case ImageStatusOrphaned:
		return "orphaned"
	case ImageStatusBlocked:
		return "blocked"
	default:
		return "unknown"
	}
}

package notifications

type ScaleNotificationConfig interface {
	Validate() error
}

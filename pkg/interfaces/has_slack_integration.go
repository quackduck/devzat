package interfaces

type hasSlackIntegration interface {
	IsOfflineSlack() bool
	GetSendToSlackChan() chan string
	GetMsgsFromSlack()
	ReplaceSlackEmoji(string) string
}

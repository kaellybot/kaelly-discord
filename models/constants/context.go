package constants

type ContextKey int

const (
	ContextKeyChannel ContextKey = iota
	ContextKeyCity
	ContextKeyDate
	ContextKeyDuration
	ContextKeyEnabled
	ContextKeyFeed
	ContextKeyJob
	ContextKeyLanguage
	ContextKeyLevel
	ContextKeyMap
	ContextKeyOrder
	ContextKeyQuery
	ContextKeyServer
	ContextKeyTwitter
)

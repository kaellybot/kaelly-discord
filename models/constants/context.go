package constants

type ContextKey int

const (
	ContextKeyChannel ContextKey = iota
	ContextKeyCity
	ContextKeyDate
	ContextKeyDimension
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

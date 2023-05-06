package constants

type ContextKey int

const (
	ContextKeyChannel ContextKey = iota
	ContextKeyCity
	ContextKeyDimension
	ContextKeyEnabled
	ContextKeyFeed
	ContextKeyJob
	ContextKeyLanguage
	ContextKeyLevel
	ContextKeyOrder
	ContextKeyServer
)

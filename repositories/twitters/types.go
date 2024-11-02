package twitters

import (
	"github.com/kaellybot/kaelly-discord/models/entities"
	"github.com/kaellybot/kaelly-discord/utils/databases"
)

type Repository interface {
	GetTwitterAccounts() ([]entities.TwitterAccount, error)
}

type Impl struct {
	db databases.MySQLConnection
}

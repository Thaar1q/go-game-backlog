package service

import (
	"errors"

	"game_backlog_api/app/repository"
	"game_backlog_api/helper"
)

// 1. Repository Error Translator
func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " not found")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict(entity + " already exists")
	default:
		return helper.Internal(err)
	}
}

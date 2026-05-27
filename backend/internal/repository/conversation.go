package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationRepo struct {
	db *pgxpool.Pool
}

func NewConversationRepo(db *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{db: db}
}

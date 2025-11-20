package repository

import (
	"context"
	"fmt"
)

type UserRepository struct {
	// db connection
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) UpdateActive(ctx context.Context, userID string, isActive bool) error {
	fmt.Println("work!")
	return nil
}

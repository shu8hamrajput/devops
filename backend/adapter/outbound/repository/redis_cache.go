package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(client *redis.Client) outbound.Cache {
	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

// SetRefreshToken stores a refresh token in Redis
func (r *RedisCache) SetRefreshToken(token string, data *entity.RefreshTokenData) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	ttl := time.Until(data.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	err = r.client.Set(r.ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves refresh token data from Redis
func (r *RedisCache) GetRefreshToken(token string) (*entity.RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var data entity.RefreshTokenData
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis
func (r *RedisCache) DeleteRefreshToken(token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

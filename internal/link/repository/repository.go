package repository

import (
	"errors"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var (
	// ideally we would define this error on a place where all repos can share a NotFound error
	// but since the scope of this shortener is very short(lol), lets keep it this way
	ErrNotFound = errors.New("repository: resource not found")
	ErrConflict = errors.New("repository: resource already exists")
)

const linkprefix = "shortener:link:"
const ipSetPrefix = "shortener:ipLinks:"

type RedisRepository struct {
	client *redis.Client
	rs     *redsync.Redsync
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	pool := goredis.NewPool(client)
	return &RedisRepository{client: client, rs: redsync.New(pool)}
}

package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyReserved = errors.New("ticket already reserved")
	ErrInvalidTTL      = errors.New("ttl must be greater than zero")
)

type Store struct {
	rdb       *redis.Client
	keyPrefix string
}

func NewStore(client *Client) *Store {
	return &Store{
		rdb:       client.Raw(),
		keyPrefix: "ticket:reservation",
	}
}

func (s *Store) key(ticketID int64) string {
	return fmt.Sprintf("%s:%d", s.keyPrefix, ticketID)
}

func (s *Store) Reserve(ctx context.Context, ticketID, bookingID int64, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTTL
	}

	key := s.key(ticketID)
	value := strconv.FormatInt(bookingID, 10)

	ok, err := s.rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrAlreadyReserved
	}

	return nil
}

var releaseScript = redis.NewScript(
	"if redis.call('GET', KEYS[1]) == ARGV[1] then " +
		"return redis.call('DEL', KEYS[1]) " +
		"else return 0 end",
)

func (s *Store) Release(ctx context.Context, ticketID, bookingID int64) (bool, error) {
	key := s.key(ticketID)
	value := strconv.FormatInt(bookingID, 10)

	n, err := releaseScript.Run(ctx, s.rdb, []string{key}, value).Int()
	if err != nil {
		return false, err
	}

	return n == 1, nil
}

func (s *Store) IsReserved(ctx context.Context, ticketID, bookingID int64) (bool, error) {
	key := s.key(ticketID)
	value := strconv.FormatInt(bookingID, 10)

	v, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return v == value, nil
}

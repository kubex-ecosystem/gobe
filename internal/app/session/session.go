// Package session fornece funções para manipulação de sessões.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Load(ctx context.Context, guild, ch, user string) (*State, error)
	Save(ctx context.Context, s *State, ttl time.Duration) error
}

type RedisStore struct{ R *redis.Client }

func key(guild, ch, user string) string { return fmt.Sprintf("KBX:session:%s:%s:%s", guild, ch, user) }

func (s RedisStore) Load(ctx context.Context, guild, ch, user string) (*State, error) {
	b, err := s.R.Get(ctx, key(guild, ch, user)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s RedisStore) Save(ctx context.Context, st *State, ttl time.Duration) error {
	st.UpdatedAtUnix = time.Now().Unix()
	b, _ := json.Marshal(st)
	return s.R.Set(ctx, key(st.GuildID, st.ChannelID, st.UserID), b, ttl).Err()
}

type State struct {
	ID              string `json:"id"`
	GuildID         string `json:"guild_id"`
	ChannelID       string `json:"channel_id"`
	UserID          string `json:"user_id"`
	LastBotState    string `json:"last_bot_state"` // IDLE|WORKING|PENDING|DONE
	SnapshotSHA     string `json:"snapshot_sha"`
	Model           string `json:"model"`
	ChunkSize       int    `json:"chunk_size"`
	ReduceURI       string `json:"reduce_uri"`
	ProgressPct     int    `json:"progress_pct"`
	NextStep        string `json:"next_step"` // MAP:chunk_007 | REDUCE | DONE
	ContextNugget   string `json:"context_nugget"`
	LastUserIntent  string `json:"last_user_intent"`
	LastMessageHash string `json:"last_message_hash"`
	LastMessageUnix int64  `json:"last_message_unix"`
	UpdatedAtUnix   int64  `json:"updated_at_unix"`
}

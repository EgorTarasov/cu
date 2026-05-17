package timeclient

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	scanInitBuf = 64 * 1024
	scanMaxBuf  = 1024 * 1024
)

// State persists incremental sync progress for a channel.
type State struct {
	ChannelID    string `json:"channel_id"`
	LastPostID   string `json:"last_post_id"`
	LastCreateAt int64  `json:"last_create_at"`
	SyncedAt     int64  `json:"synced_at"`
	BotUserID    string `json:"bot_user_id,omitempty"`
	BotUsername  string `json:"bot_username,omitempty"`
}

// Store is an append-only JSONL store for posts plus a small JSON state file.
type Store struct {
	dir string
}

func NewStore(channelID string) (*Store, error) {
	dir, err := StorageDir(channelID)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) Dir() string       { return s.dir }
func (s *Store) postsPath() string { return filepath.Join(s.dir, "posts.jsonl") }
func (s *Store) statePath() string { return filepath.Join(s.dir, "state.json") }

// LoadState reads state.json or returns a zero-value State if missing.
func (s *Store) LoadState() (State, error) {
	var st State
	data, err := os.ReadFile(s.statePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return st, nil
		}
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, fmt.Errorf("decode state: %w", err)
	}
	return st, nil
}

// SaveState writes state.json atomically.
func (s *Store) SaveState(st State) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.statePath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.statePath())
}

// AppendPosts writes posts as JSONL, ordered by CreateAt ascending.
// Returns the number of new posts persisted and the latest seen create_at/id.
func (s *Store) AppendPosts(posts []*Post) (int, string, int64, error) {
	if len(posts) == 0 {
		return 0, "", 0, nil
	}

	known, err := s.loadKnownIDs()
	if err != nil {
		return 0, "", 0, err
	}

	sort.Slice(posts, func(i, j int) bool { return posts[i].CreateAt < posts[j].CreateAt })

	f, err := os.OpenFile(s.postsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return 0, "", 0, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	count := 0
	var lastID string
	var lastAt int64
	for _, p := range posts {
		if p == nil || known[p.ID] {
			continue
		}
		line, err := json.Marshal(p)
		if err != nil {
			return count, lastID, lastAt, err
		}
		if _, err := w.Write(append(line, '\n')); err != nil {
			return count, lastID, lastAt, err
		}
		known[p.ID] = true
		count++
		lastID = p.ID
		lastAt = p.CreateAt
	}
	if err := w.Flush(); err != nil {
		return count, lastID, lastAt, err
	}
	return count, lastID, lastAt, nil
}

func (s *Store) loadKnownIDs() (map[string]bool, error) {
	out := map[string]bool{}
	f, err := os.Open(s.postsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)
	for sc.Scan() {
		var p Post
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			continue
		}
		if p.ID != "" {
			out[p.ID] = true
		}
	}
	return out, sc.Err()
}

// ReadLastN returns the last n posts from storage ordered by CreateAt ascending.
func (s *Store) ReadLastN(n int) ([]Post, error) {
	f, err := os.Open(s.postsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var all []Post
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)
	for sc.Scan() {
		var p Post
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			continue
		}
		all = append(all, p)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreateAt < all[j].CreateAt })
	if n > 0 && len(all) > n {
		all = all[len(all)-n:]
	}
	return all, nil
}

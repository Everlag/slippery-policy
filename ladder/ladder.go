package ladder

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

type Character struct {
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Class      string `json:"class"`
	ID         string `json:"id"`
	Experience int64  `json:"experience"`
	Depth      struct {
		Default int `json:"default"`
		Solo    int `json:"solo"`
	} `json:"depth"`
}

type Account struct {
	Name       string `json:"name"`
	Realm      string `json:"realm"`
	Challenges struct {
		Total int `json:"total"`
	} `json:"challenges"`
	Twitch struct {
		Name string `json:"name"`
	} `json:"twitch"`
}

type Entry struct {
	Rank      int       `json:"rank"`
	Dead      bool      `json:"dead"`
	Online    bool      `json:"online"`
	Character Character `json:"character"`
	Retired   bool      `json:"retired,omitempty"`
	Account   Account   `json:"account,omitempty"`
}

type Ladder struct {
	Total       int       `json:"total"`
	CachedSince time.Time `json:"cached_since"`
	Entries     []Entry   `json:"entries"`
}

// ActiveCharacters returns all Characters in the ladder which are not
// dead or retired.
func (l *Ladder) ActiveCharacters() []Entry {
	result := make([]Entry, 0, len(l.Entries))
	for _, e := range l.Entries {
		if e.Dead || e.Retired {
			continue
		}

		result = append(result, e)
	}
	return result
}

// ReadLadder returns a Ladder decoded from the provided Reader
func ReadLadder(r io.Reader) (Ladder, error) {
	var l Ladder
	if err := json.NewDecoder(r).Decode(&l); err != nil {
		return Ladder{}, errors.Wrap(err, "decoding ladder json")
	}
	return l, nil
}

// PageCursor lets us keep track of where we are in a ladder
// traversal.
type PageCursor struct {
	Offset int
	Limit  int
}

func (c PageCursor) String() string {
	return fmt.Sprintf("count=%d;offset=%d", c.Limit, c.Offset)
}

var _ fmt.Stringer = PageCursor{}

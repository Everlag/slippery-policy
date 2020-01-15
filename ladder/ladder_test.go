package ladder

import (
	"bytes"
	"testing"

	"github.com/Everlag/slippery-policy/fixtures"
	"github.com/stretchr/testify/require"
)

func TestReadLadder(t *testing.T) {
	blob := fixtures.FixtureBytes(t, fixtures.GetLadderFixture)
	l, err := ReadLadder(bytes.NewReader(blob))
	require.NoError(t, err)

	require.NotZero(t, l.Total, "total reports some number present")
	require.NotEmpty(t, l.Entries, "decoded some entries")
	for _, e := range l.Entries {
		require.NotEmpty(t, e.Character.Name,
			"character did not have name: %v", e)
		require.NotEmpty(t, e.Account.Name,
			"account did not have name: %v", e)
	}
}

func TestActiveCharacters(t *testing.T) {
	getFixtureLadder := func(entries ...Entry) Ladder {
		return Ladder{
			Entries: entries,
		}
	}

	t.Run("filters out dead characters", func(t *testing.T) {
		l := getFixtureLadder(Entry{
			Dead: true,
		})

		require.Empty(t, l.ActiveCharacters())
	})

	t.Run("filters out retired characters", func(t *testing.T) {
		l := getFixtureLadder(Entry{
			Retired: true,
		})

		require.Empty(t, l.ActiveCharacters())
	})

	t.Run("doesn't filter out active characters", func(t *testing.T) {
		l := getFixtureLadder(Entry{})

		require.NotEmpty(t, l.ActiveCharacters())
	})
}

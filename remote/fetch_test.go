// +build integration

package remote

import (
	"bytes"
	"testing"
	"time"

	"github.com/Everlag/slippery-policy/items"
	"github.com/Everlag/slippery-policy/ladder"
	"github.com/Everlag/slippery-policy/passives"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFetchLadder(t *testing.T) {
	cursor := ladder.PageCursor{
		Limit:  95,
		Offset: 10,
	}
	ladderName := "Slippery Hobo League (PL5357)"

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	l := NewLimiter(time.Millisecond*1500, time.Second*2,
		5, logger.With(zap.String("limiter", "snapshot")))

	result, err := FetchLadder(logger, l, cursor, ladderName)
	require.NoError(t, err)

	parsed, err := ladder.ReadLadder(bytes.NewReader(result))
	require.NoError(t, err)

	require.NotEmpty(t, parsed.Entries)
}

func TestFetchCharacter(t *testing.T) {
	logger, err := zap.NewProduction()
	require.NoError(t, err)

	characterName := "SleeperSpectreBoi"
	accountName := "Everlag"
	l := NewLimiter(time.Millisecond*1500, time.Second*2,
		5, logger.With(zap.String("limiter", "snapshot")))
	result, err := FetchCharacter(logger, l, accountName, characterName)
	require.NoError(t, err)

	parsed, err := items.ReadGetItemResp(bytes.NewReader(result))
	require.NoError(t, err)

	require.NotEmpty(t, parsed.Items)
}

func TestFetchPassives(t *testing.T) {
	logger, err := zap.NewProduction()
	require.NoError(t, err)

	characterName := "SleeperSpectreBoi"
	accountName := "Everlag"
	result, err := FetchPassives(logger, accountName, characterName)
	require.NoError(t, err)

	parsed, err := passives.ReadPassives(bytes.NewReader(result))
	require.NoError(t, err)

	require.NotEmpty(t, parsed.Items)
}

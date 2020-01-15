package remote

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/Everlag/slippery-policy/ladder"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Stolen mostly from poe-diff

func TestRateLimitHeaderName(t *testing.T) {
	t.Run("match ip state", func(t *testing.T) {
		require.True(t,
			RateLimitStateName.MatchString("X-Rate-Limit-Ip-State"))
	})

	t.Run("match account state", func(t *testing.T) {
		require.True(t,
			RateLimitStateName.MatchString("X-Rate-Limit-Account-State"))
	})

	t.Run("don't match policy", func(t *testing.T) {
		require.False(t,
			RateLimitStateName.MatchString("X-Rate-Limit-Policy"))
	})

	t.Run("don't match unprefixed", func(t *testing.T) {
		require.False(t,
			RateLimitStateName.MatchString("X-Rate-Limit-Account"))
	})

	t.Run("don't match rules", func(t *testing.T) {
		require.False(t,
			RateLimitStateName.MatchString("X-Rate-Limit-Rules"))
	})
}

func TestRateLimitHeaderState(t *testing.T) {
	t.Run("single tuple", func(t *testing.T) {
		state, err := rateLimitStateValue("3:60:0")
		require.NoError(t, err)
		require.NotZero(t, state)
		require.Equal(t, RateLimitState{
			Current: 3, Max: 60,
		}.Rel(), state.Rel())
	})

	t.Run("multiple tuples", func(t *testing.T) {
		state, err := rateLimitStateValue("3:60:0,3:240:0")
		require.NoError(t, err)
		require.NotZero(t, state)
		require.Equal(t, RateLimitState{
			Current: 3, Max: 60,
		}.Rel(), state.Rel())
	})

	t.Run("higher rel", func(t *testing.T) {
		state, err := rateLimitStateValue("3:60:0,200:240:0")
		require.NoError(t, err)
		require.NotZero(t, state)
		require.Equal(t, RateLimitState{
			Current: 200, Max: 240,
		}.Rel(), state.Rel())
	})
}

func TestLimiter(t *testing.T) {
	backoff := time.Millisecond * 20
	getLimiter := func() *Limiter {
		return NewLimiter(time.Microsecond, backoff, 0, zap.NewNop())
	}

	noBackoff := func() (RateLimitState, error) {
		return RateLimitState{
			Current: 30,
			Max:     999,
		}, nil
	}

	forceBackoff := func() (RateLimitState, error) {
		return RateLimitState{
			Current: 30,
			Max:     30,
		}, nil
	}

	t.Run("steady state", func(t *testing.T) {
		l := getLimiter()

		var maxDelta time.Duration
		for i := 0; i < 20; i++ {
			start := time.Now()
			l.Run(noBackoff)
			end := time.Now()

			delta := end.Sub(start)
			t.Log("delta is", delta)
			if delta > maxDelta {
				maxDelta = delta
			}
		}

		require.False(t, maxDelta >= backoff, "engaged in backoff")
	})

	t.Run("constant backoff", func(t *testing.T) {
		l := getLimiter()
		l.Run(forceBackoff)

		// Get around the flakiness of the scheduler by requiring that
		// we backoff at least once during a trial.
		for i := 0; i < 20; i++ {
			start := time.Now()
			l.Run(forceBackoff)
			end := time.Now()

			delta := end.Sub(start)
			if delta >= backoff {
				return
			}
		}
		t.Fatal("did not backoff during trials")
	})

	t.Run("backoff to steady state", func(t *testing.T) {
		l := getLimiter()
		l.Run(forceBackoff)

		// Get around the flakiness of the scheduler by forcing
		// backoff state to propagate thoroughly
		for i := 0; i < 5; i++ {
			l.Run(forceBackoff)
		}

		var minDelta time.Duration
		for i := 0; i < 20; i++ {
			start := time.Now()
			l.Run(noBackoff)
			end := time.Now()

			delta := end.Sub(start)
			t.Log("delta is", delta)
			if delta < minDelta {
				minDelta = delta
			}
		}
		require.False(t, minDelta >= backoff, "engaged in backoff after steadying")
	})
}

func TestLadderURL(t *testing.T) {
	cursor := ladder.PageCursor{
		Limit:  95,
		Offset: 10,
	}
	ladderName := "some-ladder (ABC12020)"

	result, err := ladderURL(cursor, ladderName)
	require.NoError(t, err)

	resultString := result.String()

	require.Contains(t, resultString,
		fmt.Sprintf("limit=%d", cursor.Limit))
	require.Contains(t, resultString,
		fmt.Sprintf("offset=%d", cursor.Offset))
	require.Contains(t, resultString, url.PathEscape(ladderName))

	// Ensure the end result is a valid url
	_, err = url.Parse(resultString)
	require.NoError(t, err)
}

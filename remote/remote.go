package remote

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Everlag/slippery-policy/ladder"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RateLimitStateName is the regexp we use to match rate limit state info
var RateLimitStateName = regexp.MustCompile(`X-Rate-Limit-.*-State`)

// BackoffThreshold is the threshold of limit saturation at which we start to
// backoff. A higher threshold is more aggressive.
const BackoffThreshold float32 = 0.6

// LimiterWindow is TODO why?
var LimiterWindow = time.Second * 30

// ErrPrivateProfile is returned when a FetchCharacter receives a 403
var ErrPrivateProfile = fmt.Errorf("got 403; profile is likely private")

// Limiter allows execution of arbitrary functions rate-limited
type Limiter struct {
	backoff time.Duration

	ticker *time.Ticker
	signal chan struct{}

	halt chan struct{}

	state           RateLimitState
	lastStateChange time.Time

	log *zap.Logger

	sync.RWMutex
}

// NewLimiter returns a Limiter which allows bursting behavior up a the specific limit.
//
// tick represents the minimum period between Takes for the Limiter.
// backoff is the period of time backoff should last. This should
// be longer than a single request.
// Rate-limiting is performed based off of received RateLimiteState.
func NewLimiter(tick, backoff time.Duration, burst int,
	logContext *zap.Logger) *Limiter {

	idBytes := make([]byte, 4)
	rand.Read(idBytes)
	id := hex.EncodeToString(idBytes)

	l := &Limiter{
		backoff: backoff,
		ticker:  time.NewTicker(tick),
		signal:  make(chan struct{}, burst),
		halt:    make(chan struct{}),
		log:     logContext.With(zap.String("Limiter", id)),
	}

	go l.start()

	return l
}

// start begins execution of the Limiter
func (l *Limiter) start() {
	for {
		select {
		case <-l.ticker.C:
			l.RWMutex.RLock()
			snapshot := l.state
			l.RWMutex.RUnlock()

			// Check if we need to backoff for a threshold
			if snapshot.Rel() >= BackoffThreshold {
				// Determine how many units we need to wait to get
				// under our threshold.
				maxCurrent := (float64(snapshot.Max) * float64(BackoffThreshold))
				waitUnits := maxCurrent - float64(snapshot.Current)

				waitTime := l.backoff * time.Duration(math.Abs(waitUnits))

				l.log.Debug("backing off",
					zap.Duration("waitTime", waitTime))

				time.Sleep(waitTime)
			}

			l.signal <- struct{}{}
		case <-l.halt:
			l.log.Debug("start exiting off halted Limiter")
			close(l.signal)
			return
		}
	}
}

// Backoff instructs the Limiter to immediately start limiting as though
// the provided state, that would trigger rate-limiting, was received.
func (l *Limiter) Backoff(s RateLimitState) {
	l.log.Debug("manually backing off Limiter")
	l.Lock()
	defer l.Unlock()

	l.updateState(s)
}

// Halt requests a Limiter to shutdown
//
// This method is not idempotent and calling beyond the first time is unsafe.
func (l *Limiter) Halt() {
	l.log.Debug("halting Limiter")
	l.halt <- struct{}{}
}

func (l *Limiter) updateState(s RateLimitState) {
	l.Lock()
	defer l.Unlock()

	now := time.Now()

	// Check if we're ready for an update;
	// or if the new rel is higher than we've recordered.
	//
	// Sometimes the API seems flaky and returns us fractions
	if now.Sub(l.lastStateChange) > LimiterWindow ||
		s.Rel() > l.state.Rel() {

		l.state = s
	}
}

// ErrLimiterHalted is the distinguished error returned from Run
// when the backing Limiter is no longer valid.
var ErrLimiterHalted = errors.New("Limiter halted")

// Run executes the provided callback while respecting rate limiting state.
//
// This blocks until an execution slot opens up.
//
// Errors from the provided function invalidate the RateLimitState and
// are passed up.
func (l *Limiter) Run(cb func() (RateLimitState, error)) error {
	_, ok := <-l.signal
	if !ok {
		l.log.Debug("read failure off halted Limiter")
		return ErrLimiterHalted
	}
	s, err := cb()
	if err == nil {
		l.updateState(s)
	}

	return err
}

// RateLimitState describes a snapshot of a remote service's rate-limiting
// behavior.
type RateLimitState struct {
	Current int
	Max     int
}

// Rel returns the relative filled of this RateLimitState
func (s RateLimitState) Rel() float32 {
	if s.Max == 0 {
		return 0
	}
	return float32(s.Current) / float32(s.Max)
}

func rateLimitStateValue(state string) (RateLimitState, error) {

	found := false
	worst := RateLimitState{}

	portions := strings.Split(state, ",")
	for _, p := range portions {
		var s RateLimitState
		_, err := fmt.Sscanf(p, "%d:%d", &s.Current, &s.Max)
		if err != nil {
			return RateLimitState{},
				errors.Wrap(err, "parsing out RateLimitStateValues")
		}

		if s.Rel() > worst.Rel() {
			worst = s
		}
		found = true
	}
	if !found {
		return RateLimitState{}, errors.New("failed to find rate limiting header")
	}

	return worst, nil
}

// GetItemsURL is the remote endpoint to fetch character information from
const GetItemsURL = "https://www.pathofexile.com/character-window/get-items"

// GetPassivesURL is the remote endpoint to fetch passives information from
const GetPassivesURL = "https://www.pathofexile.com/character-window/get-passive-skills"

// GetLadderURL is the remote endpoint to fetch ladder information
//
// Note that the ladder name, ie Slippery%20Hobo%20League%20(PL5357),
// must be appended at the end of the path with a separating slash.
const GetLadderURL = "http://api.pathofexile.com/ladders"

// FetchCharacter resolves the GetItemsURL for a provided character under a specified account.
//
// This returns the contents of the body fetched.
func FetchCharacter(logContext *zap.Logger, l *Limiter,
	accountName, characterName string) ([]byte, error) {

	var resp *http.Response
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	err := l.Run(func() (RateLimitState, error) {
		var err error
		resp, err = http.PostForm(GetItemsURL,
			url.Values{"accountName": {accountName}, "character": {characterName}})
		if err != nil {
			return RateLimitState{}, errors.Wrap(err, "fetching remote")
		}

		if resp.StatusCode != http.StatusOK {
			// Specifically instruct the rate limiter to aggressively
			// backoff when we encounter this scenario.
			if resp.StatusCode == http.StatusTooManyRequests {
				l.Backoff(RateLimitState{Current: 2, Max: 1})
			}
			// Handle private profiles specifically
			if resp.StatusCode == http.StatusUnauthorized ||
				resp.StatusCode == http.StatusForbidden {
				return RateLimitState{},
					errors.Wrap(ErrPrivateProfile, "403 received")
			}
			return RateLimitState{},
				errors.Errorf("non-200 status code: %d", resp.StatusCode)
		}

		for h := range resp.Header {
			if !RateLimitStateName.MatchString(h) {
				continue
			}

			return rateLimitStateValue(resp.Header.Get(h))
		}

		return RateLimitState{}, errors.New("failed to find rate limit information")
	})
	if err != nil {
		if err == ErrLimiterHalted {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to run fetch")
	}

	buf := bytes.NewBuffer(make([]byte, 0, 6*1024))
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return buf.Bytes(), nil
}

func ladderURL(cursor ladder.PageCursor, ladderName string) (*url.URL, error) {

	// Make sure we're not putting anything funky into our URL
	ladderName = url.PathEscape(ladderName)

	base, err := url.Parse(GetLadderURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing base URL")
	}
	path := path.Join(base.Path, ladderName)
	endpoint, err := base.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "parsing updated path")
	}

	// Update with the query parameters we have
	endpoint.RawQuery = url.Values{
		"offset": []string{strconv.Itoa(cursor.Offset)},
		"limit":  []string{strconv.Itoa(cursor.Limit)},
	}.Encode()

	return endpoint, nil
}

// FetchLadder resolves the GetItemsURL for a provided character under a specified account.
//
// This returns the contents of the body fetched.
func FetchLadder(logContext *zap.Logger, l *Limiter,
	cursor ladder.PageCursor, ladderName string) ([]byte, error) {

	var resp *http.Response
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	ladderURL, err := ladderURL(cursor, ladderName)
	if err != nil {
		return nil, errors.Wrap(err, "computing URL")
	}
	endpoint := ladderURL.String()

	err = l.Run(func() (RateLimitState, error) {
		var err error
		resp, err = http.Get(endpoint)
		if err != nil {
			return RateLimitState{}, errors.Wrap(err, "fetching remote")
		}

		if resp.StatusCode != http.StatusOK {
			logContext.Error("fetch not OK",
				zap.Int("status code", resp.StatusCode),
				zap.String("status text", resp.Status))

			// Specifically instruct the rate limiter to aggressively
			// backoff when we encounter this scenario.
			if resp.StatusCode == http.StatusTooManyRequests {
				l.Backoff(RateLimitState{Current: 2, Max: 1})
			}
			return RateLimitState{},
				errors.Errorf("non-200 status code: %d", resp.StatusCode)
		}

		for h := range resp.Header {
			if !RateLimitStateName.MatchString(h) {
				continue
			}

			return rateLimitStateValue(resp.Header.Get(h))
		}

		return RateLimitState{}, errors.New("failed to find rate limit information")
	})
	if err != nil {
		if err == ErrLimiterHalted {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to run fetch")
	}

	buf := bytes.NewBuffer(make([]byte, 0, 6*1024))
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return buf.Bytes(), nil
}

// FetchPassives resolves the GetPassivesURL for a provided character under a specified account.
//
// This returns the contents of the body fetched.
func FetchPassives(logContext *zap.Logger,
	accountName, characterName string) ([]byte, error) {

	var resp *http.Response
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	req, err := http.NewRequest("GET", GetPassivesURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "building request")
	}

	q := req.URL.Query()
	q.Add("character", characterName)
	q.Add("accountName", accountName)
	req.URL.RawQuery = q.Encode()

	// In theory, passives requests are not rate limited.
	// If this changes, our logging should catch it... hopefully
	resp, err = http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			logContext.Error("encountered unexpected rate limiting; backing off",
				zap.Int("status code", resp.StatusCode),
				zap.String("status text", resp.Status))

			// This is a minimal delay to ensure we aren't seen as actively
			// abusive.
			time.Sleep(time.Second * 1)
		}
		// Handle private profiles specifically
		if resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden {
			return nil,
				errors.Wrap(ErrPrivateProfile, "403 received")
		}
		return nil, errors.New("non-200 status code")
	}

	buf := bytes.NewBuffer(make([]byte, 0, 6*1024))
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return buf.Bytes(), nil
}

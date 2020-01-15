package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Everlag/slippery-policy/items"
	"github.com/Everlag/slippery-policy/ladder"
	"github.com/Everlag/slippery-policy/remote"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ladderPageSize = flag.Int("ladder_page_size", 5, "how many characters to process between ladder refreshes")
var ladderName = flag.String("ladder", "Slippery Hobo League (PL5357)", "which ladder to use")
var outputFile = flag.String("o", "policy_failures.csv", "output file")

func main() {
	flag.Parse()

	logger, err := getLogger()
	if err != nil {
		fmt.Println(errors.Wrap(err, "initializing logger"))
		os.Exit(1)
	}
	logger = logger.With(zap.String("ladder", *ladderName))
	logger.Debug("booting up")

	config := enforceConfig{
		Ladder: *ladderName,
		LadderLimiter: remote.NewLimiter(time.Millisecond*1500, time.Second*2,
			5, logger.With(zap.String("limiter", "ladder"))),
		CharLimiter: remote.NewLimiter(time.Millisecond*1500, time.Second*2,
			5, logger.With(zap.String("limiter", "character"))),

		// Make a best-effort attempt at deduplicating output.
		// We only care about the first violation a Character had
		//
		// We don't want to report any characters we've seen already.
		Seen: make(map[string]struct{}, 200),
	}
	coreLoop(*ladderPageSize, logger, config)
}

func getLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	// Print output and also send to a file.
	config.OutputPaths = []string{"stdout", "watch.log"}
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	return config.Build()
}

func coreLoop(pageSize int,
	logger *zap.Logger,
	config enforceConfig) error {

	out := *outputFile
	output, err := os.OpenFile(out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "opening output file: %s", out)
	}
	defer output.Close()

	writer := csv.NewWriter(output)

	// Check if this is a new file we should
	// write a header for
	stats, err := output.Stat()
	if err != nil {
		return errors.Wrap(err, "statting output file")
	}
	if stats.Size() == 0 {
		writer.Write(items.PolicyFailureCSVHeader())
		writer.Flush()
	}

	for {
		ladderCursor := ladder.PageCursor{
			Limit:  pageSize,
			Offset: 0,
		}

		logger.Info("starting from top of ladder")

		// Iterate over the entire ladder
		hadFullPage := true
		for i := 0; hadFullPage; i++ {
			logger := logger.With(zap.String("cursor", ladderCursor.String()))

			failures, foundCount, err := enforce(logger,
				ladderCursor,
				config)
			if err != nil {
				// Ignore failed pages; everything after this
				// effectively NOPs.
				//
				// However, we CANNOT continue here as the cursor
				// management needs to happen :|
				logger.Error("failed enforcing against ladder page",
					zap.Error(err))
			}

			for _, f := range failures {
				seenKey := seenKey(f.CharacterName, f.AccountName)
				if _, ok := config.Seen[seenKey]; ok {
					continue
				}

				err := writer.Write(f.ToCSVRecord())
				if err != nil {
					logger.Error("failed writing CSV line",
						zap.Error(err))
				}

				config.Seen[seenKey] = struct{}{}
			}
			// Ensure this hits the disk
			writer.Flush()
			if writer.Error() != nil {
				logger.Error("failed flushing CSV lines",
					zap.Error(err))
				return errors.Wrap(err, "flushing CSV")
			}

			// Manage our cursor and be able to wrap.
			hadFullPage = foundCount >= pageSize
			ladderCursor.Offset += pageSize
		}

	}

}

type enforceConfig struct {
	Ladder        string
	LadderLimiter *remote.Limiter

	CharLimiter *remote.Limiter

	// Keep track of the Character's we've failed on
	Seen map[string]struct{}
}

func seenKey(character, account string) string {
	return fmt.Sprintf("%s-%s", account, character)
}

func enforce(logger *zap.Logger,
	ladderCursor ladder.PageCursor,
	config enforceConfig) ([]items.PolicyFailure, int, error) {

	ladderBuf, err := remote.FetchLadder(logger,
		config.LadderLimiter, ladderCursor, config.Ladder)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "fetching ladder page %s", ladderCursor)
	}

	l, err := ladder.ReadLadder(bytes.NewReader(ladderBuf))
	if err != nil {
		return nil, 0, errors.Wrapf(err, "decoding ladder page %s", ladderCursor)
	}

	now := time.Now()
	var failures []items.PolicyFailure
	for _, c := range l.ActiveCharacters() {
		logger := logger.With(
			zap.String("account", c.Account.Name),
			zap.String("character", c.Character.Name),
		)
		logger.Debug("checking")
		buf, err := remote.FetchCharacter(logger,
			config.CharLimiter, c.Account.Name, c.Character.Name)
		if err != nil {
			if errors.Cause(err) == remote.ErrPrivateProfile {
				failures = append(failures, items.PolicyFailure{
					Reason:        items.PolicyFailureReasonPrivateProfile,
					AccountName:   c.Account.Name,
					CharacterName: c.Character.Name,
					When:          now,
				})
				continue
			}
			logger.Info("failed to find character; may have been deleted",
				zap.Error(err))
			continue
		}

		resp, err := items.ReadGetItemResp(bytes.NewReader(buf))
		if err != nil {
			logger.Info("failed to decode character; api may have changed in a way that breaks compatibility")
			continue
		}

		f := resp.EnforceGucciHobo(now, c.Account.Name)
		if len(f) == 0 {
			continue
		}
		failures = append(failures, f...)
	}

	// Include ALL characters here, including dead
	return failures, len(l.Entries), nil
}

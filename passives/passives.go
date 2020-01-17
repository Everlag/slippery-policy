package passives

import (
	"io"
	"time"

	"github.com/Everlag/slippery-policy/items"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

// GetPassivesResp is the raw response received from the JSON get-passive-skills
// api
type GetPassivesResp struct {
	Items []items.ItemResp `json:"items"`
}

// ReadPassives attempts to convert the provided blob to a GetPassiivesResp
func ReadPassives(r io.Reader) (*GetPassivesResp, error) {
	var resp GetPassivesResp
	err := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(r).Decode(&resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse blob into response")
	}

	return &resp, nil
}

// EnforceGucciHobo ensures all facets of the GetPassivesResp is valid
// under the constraints of the 'Gucci Hobo' policy.
func (r *GetPassivesResp) EnforceGucciHobo(now time.Time,
	characterName string, characterLevel int,
	accountName string) []items.PolicyFailure {

	var failures []items.PolicyFailure

	// Note that we normally exclude characters of level 2 or lower.
	// However, Characters of that level are not able to use passive
	// tree jewels

	// Filter for policy exceptions
	for _, i := range r.Items {
		fail, failed := i.EnforceGucciHobo(now,
			characterName, characterLevel, accountName)
		if !failed {
			continue
		}
		failures = append(failures, fail)
	}

	return failures
}

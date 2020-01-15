package items

import (
	"io"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	frameTypeNormal     = 0
	frameTypeMagic      = 1
	frameTypeRare       = 2
	frameTypeUnique     = 3
	frameTypeGem        = 4
	frameTypeCurrency   = 5
	frameTypeDivination = 6
	frameTypeQuestItem  = 7
	frameTypeProphecy   = 8
	frameTypeRelic      = 9

	inventoryIDFlask = "Flask"
)

const (
	// PolicyFailureReasonItem is set as the reason for a PolicyFailure
	// when the issue is a non-unique item.
	PolicyFailureReasonItem = "NonUniqueItemPresent"
	// PolicyFailureReasonPrivateProfile is set as the reason for a PolicyFailure
	// when the issue is a private profile.
	PolicyFailureReasonPrivateProfile = "PrivateProfile"
)

// PolicyFailure are the details we surface when
// disallowed items are present
type PolicyFailure struct {
	Reason string

	ItemName  string
	ItemLevel int
	// ItemSlot is the InventoryID in ItemResp, ie Flask
	ItemSlot string

	CharacterName  string
	CharacterLevel int

	AccountName string

	When time.Time
}

// ToCSVRecord formats the PolicyFailure to be fine
// for use in a CSV.
func (f *PolicyFailure) ToCSVRecord() []string {
	return []string{
		f.Reason,
		f.ItemName,
		strconv.Itoa(f.ItemLevel),
		f.ItemSlot,
		f.CharacterName,
		strconv.Itoa(f.CharacterLevel),
		f.AccountName,
		f.When.Format(time.RFC3339),
	}
}

func ParsePolicyFailureCSV(line []string) (PolicyFailure, error) {
	itemLevel, err := strconv.Atoi(line[2])
	if err != nil {
		return PolicyFailure{}, errors.Wrap(err, "parsing itemLevel")
	}
	characterLevel, err := strconv.Atoi(line[5])
	if err != nil {
		return PolicyFailure{}, errors.Wrap(err, "parsing characterLevel")
	}
	when, err := time.Parse(time.RFC3339, line[7])
	if err != nil {
		return PolicyFailure{}, errors.Wrap(err, "parsing when")
	}
	return PolicyFailure{
		Reason:         line[0],
		ItemName:       line[1],
		ItemLevel:      itemLevel, // 2
		ItemSlot:       line[3],
		CharacterName:  line[4],
		CharacterLevel: characterLevel, // 5
		AccountName:    line[6],
		When:           when,
	}, nil
}

// PolicyFailureCSVHeader returns a CSV record that can act
// as a header for PolicyFailure.ToCSVRecord
func PolicyFailureCSVHeader() []string {
	// TODO: ensure this doesn't get out of
	// sync with PolicyFailure.ToCSVRecord
	return []string{
		"reason",
		"itemName",
		"itemLevel",
		"itemSlot",
		"characterName",
		"characterLevel",
		"accountName",
		"when",
	}
}

// ItemResp is the raw response received from the JSON get-item api
type ItemResp struct {
	Ilvl     int    `json:"ilvl"`
	Name     string `json:"name"`
	TypeLine string `json:"typeLine"`
	// FrameType tells us what rarity an item is
	FrameType int `json:"frameType"`
	// InventoryID is the socket this item is within
	//
	// We care about restricting non-flasks
	InventoryID string `json:"inventoryId"`
}

// FullName returns the name derived from name and typeline of an item.
func (i *ItemResp) FullName() string {
	builder := strings.Builder{}
	builder.Grow(len(i.Name) + 1 + len(i.TypeLine))
	if len(i.Name) > 0 {
		builder.WriteString(i.Name)

		if len(i.TypeLine) > 0 {
			builder.WriteString(" ")
		}
	}
	if len(i.TypeLine) > 0 {
		builder.WriteString(i.TypeLine)
	}
	return builder.String()
}

// ItemRespSet is a slice of ItemResp received from the get-item API
type ItemRespSet []ItemResp

// CharacterResp is a raw character representation received from the get-item API
type CharacterResp struct {
	Name            string `json:"name"`
	League          string `json:"league"`
	ClassID         int    `json:"classId"`
	AscendancyClass int    `json:"ascendancyClass"`
	Class           string `json:"class"`
	Level           int    `json:"level"`
	Experience      int64  `json:"experience"`
}

// GetItemResp is the top-level structure of a request against the get-item API
type GetItemResp struct {
	Items     ItemRespSet   `json:"items"`
	Character CharacterResp `json:"character"`
}

// EnforceGucciHobo ensures all facets of the GetItemResp is valid
// under the constraints of the 'Gucci Hobo' policy.
func (r *GetItemResp) EnforceGucciHobo(now time.Time,
	accountName string) []PolicyFailure {

	var failures []PolicyFailure

	// We have an exlusion for characters of level 2 and lower
	// since the game requires you to use a piece of equipment to
	// get past the twilight strand.
	if r.Character.Level <= 2 {
		return failures
	}

	// Filter for policy exceptions
	for _, i := range r.Items {
		// Ignore flasks
		if i.InventoryID == inventoryIDFlask {
			continue
		}

		// Allow uniques or relics, which are fancy uniques
		if i.FrameType == frameTypeUnique ||
			i.FrameType == frameTypeRelic {
			continue
		}

		failures = append(failures, PolicyFailure{
			Reason:         PolicyFailureReasonItem,
			AccountName:    accountName,
			CharacterName:  r.Character.Name,
			CharacterLevel: r.Character.Level,
			ItemName:       i.FullName(),
			ItemLevel:      i.Ilvl,
			ItemSlot:       i.InventoryID,
			When:           now,
		})
	}

	return failures
}

// ReadGetItemResp parses the provided reader as a GetItemResp
func ReadGetItemResp(r io.Reader) (*GetItemResp, error) {
	var resp GetItemResp
	err := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(r).Decode(&resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse blob into response")
	}

	return &resp, nil
}

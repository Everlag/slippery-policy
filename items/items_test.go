package items

import (
	"bytes"
	"testing"
	"time"

	"github.com/Everlag/slippery-policy/fixtures"
	"github.com/stretchr/testify/require"
)

func TestReadGetItemResp(t *testing.T) {
	blob := fixtures.FixtureBytes(t, fixtures.GetItemsFixture)
	resp, err := ReadGetItemResp(bytes.NewReader(blob))
	require.NoError(t, err)

	require.NotZero(t, resp.Character)
	require.NotEmpty(t, resp.Items)
}

func TestEnforceGucciHobo(t *testing.T) {
	const charName = "some-character"
	const accountName = "some-account"

	now, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	require.NoError(t, err, "sample time must parse")

	run := func(charLevel int, items ...ItemResp) []PolicyFailure {
		resp := GetItemResp{
			Character: CharacterResp{
				Name:  charName,
				Level: charLevel,
			},
			Items: items,
		}

		return resp.EnforceGucciHobo(now, accountName)
	}

	t.Run("happy path unique item", func(t *testing.T) {
		failures := run(99, ItemResp{
			FrameType: frameTypeUnique,
		})
		require.Empty(t, failures)
	})

	t.Run("level 2 character allowed with magic item", func(t *testing.T) {
		failures := run(2, ItemResp{
			FrameType: frameTypeMagic,
		})
		require.Empty(t, failures)
	})

	t.Run("failure reports details", func(t *testing.T) {
		badName := "some-item-name"
		badSlot := "Weapon"
		failures := run(99,
			ItemResp{
				Name:        badName,
				FrameType:   frameTypeNormal,
				InventoryID: badSlot,
				Ilvl:        84,
			},
		)
		require.NotEmpty(t, failures)

		// We ensure these are exactly equivalent as it enforces
		// that the test is updated if the code is updated.
		exactFailure := PolicyFailure{
			Reason:      PolicyFailureReasonItem,
			AccountName: accountName,

			CharacterLevel: 99,
			CharacterName:  charName,

			ItemName:  badName,
			ItemLevel: 84,
			ItemSlot:  badSlot,

			When: now,
		}
		require.Equal(t, exactFailure, failures[0])
	})

	t.Run("non-unique non-flask invalid", func(t *testing.T) {
		failures := run(99, ItemResp{
			FrameType: frameTypeMagic,
		})
		require.NotEmpty(t, failures)
	})

	t.Run("non-unique flask valid", func(t *testing.T) {
		failures := run(99, ItemResp{
			FrameType:   frameTypeMagic,
			InventoryID: inventoryIDFlask,
		})
		require.Empty(t, failures)
	})

	t.Run("multiple items succeed when all in policy", func(t *testing.T) {
		failures := run(99,
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Offhand",
			},
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Gloves",
			},
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Ring",
			},
			ItemResp{
				FrameType:   frameTypeNormal,
				InventoryID: "Flask",
			},
		)
		require.Empty(t, failures)
	})

	t.Run("multiple items fail when one not in policy", func(t *testing.T) {
		badSlot := "Weapon"
		failures := run(99,
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Offhand",
			},
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Gloves",
			},
			ItemResp{
				FrameType:   frameTypeUnique,
				InventoryID: "Ring",
			},
			ItemResp{
				FrameType:   frameTypeNormal,
				InventoryID: "Flask",
			},
			ItemResp{
				FrameType:   frameTypeNormal,
				InventoryID: badSlot,
			},
		)
		require.NotEmpty(t, failures)
	})
}

func TestPolicyFailureCSV(t *testing.T) {
	// This ensures any change to the failure MUST be explicit
	t.Run("correctly decodes", func(t *testing.T) {
		// Grab an arbitrary time to use as a reference point
		now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
		require.NoError(t, err, "parsing fixture time")

		failure := PolicyFailure{
			Reason:         "some-reason",
			ItemName:       "some-item",
			ItemLevel:      84,
			ItemSlot:       "Boots",
			CharacterName:  "Tim",
			CharacterLevel: 91,
			AccountName:    "some-account",
			When:           now,
		}

		line := failure.ToCSVRecord()

		found, err := ParsePolicyFailureCSV(line)
		require.NoError(t, err)
		require.Equal(t, failure, found)
	})
}

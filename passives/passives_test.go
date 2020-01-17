package passives

import (
	"testing"
	"time"

	"github.com/Everlag/slippery-policy/items"
	"github.com/stretchr/testify/require"
)

func TestEnforceGucciHobo(t *testing.T) {
	const charName = "some-character"
	const accountName = "some-account"

	now, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	require.NoError(t, err, "sample time must parse")

	run := func(charLevel int, items ...items.ItemResp) []items.PolicyFailure {
		resp := GetPassivesResp{
			Items: items,
		}

		return resp.EnforceGucciHobo(now, charName, charLevel, accountName)
	}

	t.Run("happy path unique item", func(t *testing.T) {
		failures := run(99, items.ItemResp{
			FrameType: items.FrameTypeUnique,
		})
		require.Empty(t, failures)
	})

	t.Run("failure reports details", func(t *testing.T) {
		badName := "some-item-name"
		badSlot := "Weapon"
		failures := run(99,
			items.ItemResp{
				Name:        badName,
				FrameType:   items.FrameTypeNormal,
				InventoryID: badSlot,
				Ilvl:        84,
			},
		)
		require.NotEmpty(t, failures)

		// We ensure these are exactly equivalent as it enforces
		// that the test is updated if the code is updated.
		exactFailure := items.PolicyFailure{
			Reason:      items.PolicyFailureReasonItem,
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
}

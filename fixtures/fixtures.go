package fixtures

import (
	"testing"

	"github.com/gobuffalo/packr"
	"github.com/stretchr/testify/require"
)

const GetItemsFixture = "get-items.json"
const GetLadderFixture = "get-ladder.json"
const EmptyItemsFixture = "lifecycle/get-items.empty.json"
const FullItemsFixture = "lifecycle/get-items.full.json"
const GetPassivesFixture34 = "get-passive-skills.json"

var box = packr.NewBox(".")

// FixtureBytes returns the bytes from a file relative to the fixtures directory
func FixtureBytes(t *testing.T, fixture string) []byte {
	if t != nil {
		t.Log(fixture)
	}
	blob, err := box.MustBytes(fixture)
	if t != nil {
		require.NoError(t, err, "reading fixture file")
	}

	return blob
}

package pob

import (
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Everlag/slippery-policy/items"
	"github.com/pkg/errors"
)

// DecodePOBCode reads in a Path of Building code from the provided
// reader and returns the decoded PathOfBuilding struct.
//
// A PoB code is a deflated-base64url-encoded XML string
// of the struct PathOfBuilding.
func DecodePOBCode(in io.Reader) (PathOfBuilding, error) {
	base64URLDecoder := base64.NewDecoder(base64.URLEncoding, in)
	zlibDecoder, err := zlib.NewReader(base64URLDecoder)
	if err != nil {
		return PathOfBuilding{}, errors.Wrap(err, "initializing nested zlib-base64 decoder")
	}
	var v PathOfBuilding
	if err := xml.NewDecoder(zlibDecoder).Decode(&v); err != nil {
		return PathOfBuilding{}, errors.Wrap(err, "decoding PoB code")
	}
	zlibDecoder.Close()

	return v, nil
}

// EncodePOBCode exports a Path of Building code from the provided
// PathOfBulding and outputs to the provided Writer
func EncodePOBCode(v PathOfBuilding, w io.Writer) error {
	// Then, base64url encode the serialized output
	base64url := base64.NewEncoder(base64.RawURLEncoding, w)
	// Finally, compress before sending to the external writer
	compressor := zlib.NewWriter(base64url)
	// Serialize our output first
	encoder := xml.NewEncoder(compressor)

	if err := encoder.Encode(v); err != nil {
		return errors.Wrap(err, "encoding xml")
	}
	if err := compressor.Close(); err != nil {
		return errors.Wrap(err, "zlib compressing PoB Code")
	}
	if err := base64url.Close(); err != nil {
		return errors.Wrap(err, "base64urling PoB Code")
	}
	return nil
}

// PathOfBuilding is raw output from jamming the xml into
// a PoB decoder.
type PathOfBuilding struct {
	XMLName xml.Name `xml:"PathOfBuilding"`
	Text    string   `xml:",chardata"`
	Build   struct {
		Text            string `xml:",chardata"`
		Level           string `xml:"level,attr"`
		TargetVersion   string `xml:"targetVersion,attr"`
		BanditNormal    string `xml:"banditNormal,attr"`
		Bandit          string `xml:"bandit,attr"`
		BanditMerciless string `xml:"banditMerciless,attr"`
		ClassName       string `xml:"className,attr"`
		AscendClassName string `xml:"ascendClassName,attr"`
		MainSocketGroup string `xml:"mainSocketGroup,attr"`
		ViewMode        string `xml:"viewMode,attr"`
		BanditCruel     string `xml:"banditCruel,attr"`
		PlayerStat      []struct {
			Text  string `xml:",chardata"`
			Stat  string `xml:"stat,attr"`
			Value string `xml:"value,attr"`
		} `xml:"PlayerStat"`
	} `xml:"Build"`
	// Import is a self-closing tag, looks like a flag.
	// TODO: does this matter?
	// Import string `xml:"Import"`
	Skills struct {
		Text                string `xml:",chardata"`
		DefaultGemQuality   string `xml:"defaultGemQuality,attr"`
		DefaultGemLevel     string `xml:"defaultGemLevel,attr"`
		ShowSupportGemTypes string `xml:"showSupportGemTypes,attr"`
		SortGemsByDPS       string `xml:"sortGemsByDPS,attr"`
		Skill               []struct {
			Text                 string `xml:",chardata"`
			MainActiveSkillCalcs string `xml:"mainActiveSkillCalcs,attr"`
			Enabled              string `xml:"enabled,attr"`
			Slot                 string `xml:"slot,attr"`
			MainActiveSkill      string `xml:"mainActiveSkill,attr"`
			Source               string `xml:"source,attr"`
			Label                string `xml:"label,attr"`
			Gem                  []struct {
				Text          string `xml:",chardata"`
				EnableGlobal2 string `xml:"enableGlobal2,attr"`
				Quality       string `xml:"quality,attr"`
				Level         string `xml:"level,attr"`
				GemID         string `xml:"gemId,attr"`
				SkillID       string `xml:"skillId,attr"`
				EnableGlobal1 string `xml:"enableGlobal1,attr"`
				Enabled       string `xml:"enabled,attr"`
				NameSpec      string `xml:"nameSpec,attr"`
			} `xml:"Gem"`
		} `xml:"Skill"`
	} `xml:"Skills"`
	Tree struct {
		Text       string `xml:",chardata"`
		ActiveSpec string `xml:"activeSpec,attr"`
		Spec       struct {
			Text        string `xml:",chardata"`
			TreeVersion string `xml:"treeVersion,attr"`
			URL         string `xml:"URL"`
			Sockets     struct {
				Text   string `xml:",chardata"`
				Socket []struct {
					Text   string `xml:",chardata"`
					NodeID string `xml:"nodeId,attr"`
					ItemID string `xml:"itemId,attr"`
				} `xml:"Socket"`
			} `xml:"Sockets"`
		} `xml:"Spec"`
	} `xml:"Tree"`
	Notes    string `xml:"Notes"`
	TreeView struct {
		Text                string `xml:",chardata"`
		SearchStr           string `xml:"searchStr,attr"`
		ZoomY               string `xml:"zoomY,attr"`
		ShowHeatMap         string `xml:"showHeatMap,attr"`
		ZoomLevel           string `xml:"zoomLevel,attr"`
		ShowStatDifferences string `xml:"showStatDifferences,attr"`
		ZoomX               string `xml:"zoomX,attr"`
	} `xml:"TreeView"`
	ItemsUnion ItemsUnion `xml:"Items"`
}

type ItemsUnion struct {
	Text               string  `xml:",chardata"`
	ActiveItemSet      string  `xml:"activeItemSet,attr"`
	UseSecondWeaponSet string  `xml:"useSecondWeaponSet,attr"`
	Item               []Item  `xml:"Item"`
	Slot               []Slot  `xml:"Slot"`
	ItemSet            ItemSet `xml:"ItemSet"`
}

type Item struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
}

type Slot struct {
	Text   string `xml:",chardata"`
	Name   string `xml:"name,attr"`
	ItemID string `xml:"itemId,attr"`
	Active string `xml:"active,attr"`
}

type ItemSet struct {
	Text               string `xml:",chardata"`
	UseSecondWeaponSet string `xml:"useSecondWeaponSet,attr"`
	ID                 string `xml:"id,attr"`
	Slot               []Slot `xml:"Slot"`
}

func ItemRespSetToItemsUnion(r items.ItemRespSet) ItemsUnion {

	items := make([]Item, 0, len(r))
	slots := make([]Slot, 0, len(r))
	for id, i := range r {
		var itemText strings.Builder
		writeLine := func(format string, a ...interface{}) {
			itemText.WriteString(fmt.Sprintf(format, i.Ilvl))
			itemText.WriteString("\n")
		}
		writeLine(i.Name)
		writeLine(i.TypeLine)
		writeLine(fmt.Sprintf("Item Level: %d", i.Ilvl))
		// TODO: correctly handle implicits, encluding enchants
		// which, in PoB schema, are included in the implicits
		// count but are crafted...
		writeLine("Implicits: 0")

		for _, m := range i.EnchantMods {
			writeLine("{crafted}%s", m)
		}
		for _, m := range i.ImplicitMods {
			writeLine("%s", m)
		}
		for _, m := range i.UtilityMods {
			writeLine("%s", m)
		}
		for _, m := range i.ExplicitMods {
			writeLine("%s", m)
		}
		for _, m := range i.CraftedMods {
			writeLine("{crafted}%s", m)
		}

		idString := strconv.Itoa(id)
		outItem := Item{
			Text: itemText.String(),
			ID:   idString,
		}
		items = append(items, outItem)

		slotName := i.InventoryID
		if slotName == "Flask" {
			slotName = fmt.Sprintf("Flask %d", i.X)
		}
		// PoB specific translation for rings
		if slotName == "Ring2" {
			slotName = fmt.Sprintf("Ring 2")
		}

		outSlot := Slot{
			Active: "true",
			Name:   slotName,
			ItemID: idString,
		}
		slots = append(slots, outSlot)
	}

	activeSlot := "1"
	out := ItemsUnion{
		ActiveItemSet:      activeSlot,
		UseSecondWeaponSet: "false",

		Item: items,
		Slot: slots,
		ItemSet: ItemSet{
			ID:   activeSlot,
			Slot: slots,
		},
	}

	return out
}

package pob

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/Everlag/slippery-policy/items"
	"github.com/pkg/errors"
)

// SeedExportCode is used to create a baseline PathOfBuilding
// that won't crash the program when imported.
//
// The zero-value of PathOfBuilding WILL crash Path of Building.
//
// NOTE: this version is from 1.4.157.1, specifically the community fork.
// If PoB starts refusing our codes, we SHOULD regenerate this.
// const SeedExportCode = "eNqtWVt32jgQfk5_hQ_vCTEXB3pIewi5lHNCysZpu_vUo5gB1MiS15JD6a_fkWwHNwlhTLcP1Ja_-WY0o7nYGXz8GQvvEVLNlTxt-EfHDQ9kpGZcLk4bX-4uD3uNjx_eDabMLD_PzzIu7JMP7w4G7toT8AgC5RqeYekCzNeSqf0dme6ZnHFzo9KYIehGSSjXJpBGXIDW5XIkmNY3LIbTRhghQ8NjOgI5G23Wc2DMuAxV9ADmKlVZ4nQnTJolKDnhUqVXalZin9bZj9_WHzmsJmqGnOPJ9PPtXWnVKM2gNBT3eDCYCraGNDTMeBp_ThtDdBVbwDmL8ReJmMiQpX3U6zSar0uECcDsCekftf4YOE3hYj6HyPBHGKVo9ZLJaGPM8Ta5uthJJgxPBIe0YlV3m8SnF-T9YBv2ThkmzqfhE7RzFBz33kYrs9vqb9wsp4prJcnkEybZSGkCeWg2fmhtRd3Cv1XgVtw5_KSxVYFbcWNpaGxV4FbcNZ9vohgcv3Fgo_cWO5YRjfSLTEFD-lg55wFRYIoFAyrG-8dvCt7CAiTNqGuAaHmFVeWWGUJm2APzhOq-7RyLJTnHAl9xTpcoUMM5VvCZc476b4FruudCQrpYh0sOYkbIKeulqgjJW1UBYqCrInV39Mi0a0hlUHpv7yaH08IOAgAFZvCsdLZ6_6PINFU_bLcQNeWGaayylBjEHEza9XS51jzCku6a6C3MMjROUYJYdr2JeoQYD7xrmdjINwd_m-iZwJGB2vyQVohaEkNjWPRwrmYLqKWklsQlT9Fbmlea1eH28jlSNjmo6Gu-WBqJo10NBUumNB2-sf4zzlAjlhCGkKctkEWe7YOuarMZssxUrRC5tHOvroeesE1Pb291WAry15rM_xucpOBCzrLUHj2yjucSL9UMmu69wF6N40Slxi2OmIi0oxzLJDOedPO8fuBCfJdZfG8nzPz_TQpXkTHX0ff7bD63Q3sDjUnd68nF5eXF6G789aIQCcEVEi9SQrBEA877kouGx_EitMpCLJ6RIaBxmC1G_N1YO5kS9LvpfjfODt40A4uyQdz63ToBG0lNEDgTNFsxwkxQLMhncgrjQnJDcjpEbE3APfV6AvZCwJAL21coTprgmcwbEgWNDSLl95kBCtjN3wQL7CRK2FZ1IiNYmnd9Am8xF-1GFk2CEFWXc3fsASQxAucwB8wCXRSfstAM3JnX3gzmDN9iryD-K2OCm3Uhvlm_zj9guFW9VKswS2zZwic2YzT64_oan-RL-myNL5anDZNmeYAGdymAx9xQYscgW75yu_HGM_iw8kGk7x4dDL7cXruLg6UxiX7fbK5Wq6OEmaWaw0-c0o4iFTcTpjWSHroieWiZmkP8d2Z_hqenjqhZMg3yDyI6py3uPInFcox-OuljV_Ywr2J7VxT3l7B26yRoE3CtXuekS8EFfj-g6A2CdoeCa_d7fYpev9-j8HWCXq9FwAV-x6fo7Xb81gmFr9WmhKPb9fsUXKd3EvQobun0fdJue6Ro4GZJpyU4Pml3SdEN2hT72p1Oj6K37QckXCs4ab1i36BZSSm8wXR2Bcbmu724UVjK7TO7WN64avCVw8rTwNJoGRqcZxreL6Xifyy1KzCfgJkJS4qSY58VJahdFCBs0uccG0vqCltRbBzw79K-wRjN1UXdsdchGPchNNOAFVPJ2TdgiZJu2erJJ6ocuBXkiqpfFKlQqHL6uhL4uqVf8VEFk1N5PgnVClcs8Yb3a62Z8IqAtN4WPQNhaopcCqYfvPYOXjVbe3nPq0k_jDMBhmLCH3jFp_B367uOxLvDAZ8Ah6W6QSk369nd0s5UTdtLDeS4U1zR2c9UqjP2EvMJPrzFF6XdrlBG1zQhrwl1c3h7rvn7qPP3OZ_-fiHxCT6kRW2PNHdR9PenJyXyHofE3114XiAGzaIPuX7qGpn7RKDknC8QMWg-_1vjf61p0MU="
const SeedExportCode = "eNqtWVt32jgQfk5_hQ_vCTEXB3qgPYRcyjmQsnHa7j71KGYANULyWnIo_fU7ku3gJiGM6faB2vI334xmNBc7vY8_V8J7hERzJfs1_-S05oGM1IzLRb_25e7quFP7-OFdb8rM8vP8POXCPvnw7qjnrj0BjyBQruYZlizAfC2Ymt-R6Z7JGTc3KlkxBN0oCcVav1ZcTSCJuACtC0AkmNY3bAX9WhghV81jOgI5G27XM-CKcRmq6AHMdaLS2FkRM2mWoOSES5Vcq1mBfVpnP35bf-SwnqgZco4m08-3d4VVwySFwmTc7VFvKtgGktAw42n86dcG6DS2gAu2wl8kYiJFluZJp1Wrvy4RxgCzJ6R_0vhj4DSBy_kcIsMfYZig1Usmo60xp7vkqmInqTA8FhySklXtXRKfXpB3g13YO2WYuJiGT9DWSXDaeRutzH6rv3GznCqulSSTT5hkQ6UJ5KHZ-qGxE3UL_5aBO3EX8JPGVgbuxI2kobGVgTtxYz7fRjE4fePARu8tdiQjGukXmYCG5LF0zgOiwBQLBpSM90_fFLyFBUiaUWOAaHmNVeWWGUJm2APzhGq_7RyLJTnHAl9xTpsoUME5VvCZc066b4EruudSQrLYhEsOYkbIKeulsgjJW2UBYqDLIlV39Mi0a0hFUDpv7yaD08IOAgAFZvCsdDY6_6PINFE_bLcQFeUGyUqlCTGIGZi06-lyo3mEJd010VuYpWicogSx6HoT9QgrPPCuZWIj3x78XaLnAkcGavNDWiEqSQyMYdHDhZotoJKSShJXPEFvaV5qVse7y-dQ2eSgosd8sTQSh7wKCpZMaTp8a_1nnKGGLCYMIU9bIIs82wdd1XYzZJmpWiNyaSdgXQ09Ydue3tzpsATkrw2Z_zc4ScGlnKWJPXpkHc8lXqrp1d0bgr0arWKVGLc4ZCLSjnIk49R40s3z-oEL8V2mq3s7YWb_b1O4jFxxHX2_T-dzO7TX0JjEvahcXl1dDu9GXy9zkRBcIfEiJQSLNeC8L7moeRwvQqssxOIZGQIah9l8xN-PtZMpQb-b7vfj7OBNMzAvG8St321isJHUBIFzQbMVI8wExYJsJqcwLiQ3JKdDxDYE3FOvJ2AvBQy4sH2F4qQJnsmsIVHQ2CASfp8aoIDd_E2wwE6ihG2VJzKCpVnXJ_Dmc9F-ZN4kCFF1OXfHHkASI3ABc8As0HnxKQpNz5157c1gzvAt9hpWf6VMcLPJxbfr4-xThlvVS7UO09iWLXxiM0ajP8ZjfJIt6fMNvlj2ayZJswD17hIAj7mhxI5BtnxlduONZ_Bh6dNI1z066n25HbuLo6UxsX5fr6_X65OYmaWaw0-c0k4itarHTGskPXZF8tgy1Qf479z-DPp9R1QvmHrZBxGd0eZ3nsRiOUI_nXWxK3uYVyt7lxf3l7Bm4yxoEnCNTuusTcEFfjeg6A2CZouCa3Y7XYpev9uh8LWCTqdBwAV-y6fobbf8xhmFr9GkhKPd9rsUXKtzFnQobml1fdJuO6Ro4GZJpyU4PWu2SdENmhT7mq1Wh6K36QckXCM4a7xiX69eSim8wXR2Bcbmu724UVjK7TO7WNy4avCVw9rTwJJoGZrEfvH8pdTqH0vtCswnYGbC4rzk2Gd5CWrmBQib9AXHxpK4wpYXGwf8u7CvN0JzdV537HUIxn0ITTVgxVRy9g1YrKRbtnqyiSoD7gS5ournRSoUqpi-rgW-bulXfFTCZFSeT0I1wjWLvcH9RmsmvDwgjbdFz0GYiiJXgukHr7mHV802XtbzKtIPVqkAQzHhD7ziU_jb1V1H4t3jgE-Aw1LVoBSb9exuaWeqou2FBnLcKa5oHWYq1RkHifkEH97ii9J-VyijK5qQ1YSqObw71_xD1PmHnE__sJD4BB_SonZAmrso-ofTkxL5gEPi7y88LxC9et6HXD91jcx9IlByzheI6NWf_9XxP-R609g="

// DecodePOBCode reads in a Path of Building code from the provided
// reader and returns the decoded PathOfBuilding struct.
//
// A PoB code is a deflated-base64url-encoded XML string
// of the struct PathOfBuilding.
func DecodePOBCode(in io.Reader) (PathOfBuilding, error) {
	dec, err := XMLDecoder(in)
	if err != nil {
		return PathOfBuilding{}, errors.Wrap(err, "initializing nested zlib-base64 decoder")
	}
	var v PathOfBuilding
	if err := xml.NewDecoder(dec).Decode(&v); err != nil {
		return PathOfBuilding{}, errors.Wrap(err, "decoding PoB code")
	}
	dec.Close()

	return v, nil
}

// XMLDecoder wraps a reader to convert a PoB code to the XML from a
// Path of Building export. Note that this preserves the XML exactly
// whereas going through `DecodePOBCode` uses our internal data structure.
func XMLDecoder(in io.Reader) (io.ReadCloser, error) {
	base64URLDecoder := base64.NewDecoder(base64.URLEncoding, in)
	zlibDecoder, err := zlib.NewReader(base64URLDecoder)
	if err != nil {
		return nil, errors.Wrap(err, "initializing nested zlib-base64 decoder")
	}
	return zlibDecoder, nil
}

// NewPathOfBuilding returns a default PathOfBuilding which is
// hydrated from SeedExportCode to avoid it crashing when imported.
func NewPathOfBuilding() (PathOfBuilding, error) {
	return DecodePOBCode(bytes.NewBufferString(SeedExportCode))
}

// newlineSquisher is used to squash multiple consecutive newlines
// into a single newline.
var newlineSquisher = regexp.MustCompile(`\n+`)

var escapedCharDataEntities = map[string]string{
	"&#34;": "\"",
	"&#39;": "'",
	"&amp;": "&",
	"&#xD;": "",
	// < and > are escaped but those make a lot
	// of sense to leave escaped...
	// "&lt;": "<",
	// "&gt;": ">",
}

const escapedNewline = "&#x9;"
const escapedLF = "&#xA;"

// EncodePOBCode exports a Path of Building code from the provided
// PathOfBulding and outputs to the provided Writer
func EncodePOBCode(v PathOfBuilding, w io.Writer) error {
	// Then, base64url encode the serialized output
	base64url := base64.NewEncoder(base64.URLEncoding, w)
	// Finally, compress before sending to the external writer
	compressor := zlib.NewWriter(base64url)
	// Serialize our output first
	var tempBuf bytes.Buffer
	encoder := xml.NewEncoder(&tempBuf)
	encoder.Indent("", " ")

	if err := encoder.Encode(v); err != nil {
		return errors.Wrap(err, "encoding xml")
	}

	result := tempBuf.String()

	// Handle newlines separately
	result = strings.Replace(result, escapedNewline, "\n", -1)
	result = strings.Replace(result, escapedLF, "\n", -1)
	result = newlineSquisher.ReplaceAllString(result, "\n")
	for old, new := range escapedCharDataEntities {
		result = strings.Replace(result, old, new, -1)
	}

	// Add a header on so the headers of the codes look the same
	// as those output from pob
	if _, err := compressor.Write([]byte(xml.Header)); err != nil {
		return errors.Wrap(err, "writing header to zlib compressor")
	}
	if _, err := compressor.Write([]byte(result)); err != nil {
		return errors.Wrap(err, "writing body to zlib compressor")
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
	// Import is a self-closing tag that we want to preserve
	Import string `xml:"Import"`
	Calcs  struct {
		Text  string `xml:",chardata"`
		Input []struct {
			Text   string `xml:",chardata"`
			Name   string `xml:"name,attr"`
			String string `xml:"string,attr"`
			Number string `xml:"number,attr"`
		} `xml:"Input"`
		Section []struct {
			Text      string `xml:",chardata"`
			Collapsed string `xml:"collapsed,attr"`
			ID        string `xml:"id,attr"`
		} `xml:"Section"`
	} `xml:"Calcs"`
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
	Config     struct {
		Text  string `xml:",chardata"`
		Input []struct {
			Text    string `xml:",chardata"`
			Name    string `xml:"name,attr"`
			Boolean string `xml:"boolean,attr"`
			String  string `xml:"string,attr"`
		} `xml:"Input"`
	} `xml:"Config"`
}

// ItemsUnion contains all item-related info
type ItemsUnion struct {
	Text               string  `xml:",chardata"`
	ActiveItemSet      string  `xml:"activeItemSet,attr"`
	UseSecondWeaponSet string  `xml:"useSecondWeaponSet,attr"`
	Item               []Item  `xml:"Item"`
	Slot               []Slot  `xml:"Slot"`
	ItemSet            ItemSet `xml:"ItemSet"`
}

// Item is a PoB item where all details, apart from the ID,
// are included in a line-delimited manner in the text body.
type Item struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
}

// Slot is a reference to Item that associates it with a display-level slot.
type Slot struct {
	Text   string `xml:",chardata"`
	Name   string `xml:"name,attr"`
	ItemID string `xml:"itemId,attr"`
	Active string `xml:"active,attr"`
}

// ItemSet is a group of items that has slots
type ItemSet struct {
	Text               string `xml:",chardata"`
	UseSecondWeaponSet string `xml:"useSecondWeaponSet,attr"`
	ID                 string `xml:"id,attr"`
	Slot               []Slot `xml:"Slot"`
}

// requiredSlots are slots that MUST be initialized, even with itemId=0 for
// non-present items or PoB is liable to crash :|
var requiredSlots = []string{
	"Gloves",
	"Weapon 1",
	"Weapon 2Swap Abyssal Socket 2",
	"Belt Abyssal Socket 2",
	"Flask 3",
	"Body Armour Abyssal Socket 2",
	"Amulet",
	"Flask 1",
	"Weapon 2Swap Abyssal Socket 1",
	"Flask 5",
	"Belt Abyssal Socket 1",
	"Flask 2",
	"Helmet Abyssal Socket 2",
	"Weapon 2 Swap",
	"Weapon 1 Abyssal Socket 1",
	"Weapon 2",
	"Body Armour",
	"Flask 4",
	"Weapon 1 Abyssal Socket 2",
	"Weapon 2 Abyssal Socket 2",
	"Weapon 1 Swap",
	"Ring 2",
	"Boots Abyssal Socket 2",
	"Gloves Abyssal Socket 2",
	"Body Armour Abyssal Socket 1",
	"Gloves Abyssal Socket 1",
	"Helmet Abyssal Socket 1",
	"Weapon 2 Abyssal Socket 1",
	"Boots",
	"Weapon 1Swap Abyssal Socket 1",
	"Ring 1",
	"Weapon 1Swap Abyssal Socket 2",
	"Helmet",
	"Boots Abyssal Socket 1",
	"Belt",
	// Not sure what these are but PoB crashes without them :|
	"Weapon 3",
}

// ItemRespSetToItemsUnion converts an API response to a ItemsUnion
// suitable for PoB output
func ItemRespSetToItemsUnion(r items.ItemRespSet) ItemsUnion {

	itemOut := make([]Item, 0, len(r))
	slotOut := make([]Slot, 0, len(r))
	for id, i := range r {
		var itemText strings.Builder
		writeLine := func(format string, a ...interface{}) {
			// Ignore things that would cause useless newlines
			if len(format) == 0 {
				return
			}
			itemText.WriteString(fmt.Sprintf(format, a...))
			itemText.WriteString("\n")
		}
		var rarity string
		switch i.FrameType {
		case items.FrameTypeNormal:
			rarity = "NORMAL"
		case items.FrameTypeMagic:
			rarity = "MAGIC"
		case items.FrameTypeRare:
			rarity = "RARE"
		case items.FrameTypeUnique:
			rarity = "UNIQUE"
		case items.FrameTypeRelic:
			rarity = "RELIC"
		default:
			// Fallthrough; not great but unexpected :|
			rarity = "NORMAL"
		}
		writeLine("			Rarity: %s", rarity)
		writeLine(i.Name)
		writeLine(i.TypeLine)
		writeLine(fmt.Sprintf("Item Level: %d", i.Ilvl))
		// Implicits are easy to handle; there's no penalty for getting
		// this wrong apart from the display being messed up.
		//
		// Enchants count as crafted implicits.
		writeLine("Implicits: %d",
			len(i.EnchantMods)+len(i.ImplicitMods))

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

		// Start indexing at 1 rather than 0 as 0 is reserved
		// for 'not present'
		idString := strconv.Itoa(id + 1)
		outItem := Item{
			Text: itemText.String(),
			ID:   idString,
		}
		itemOut = append(itemOut, outItem)

		slotName := i.InventoryID
		// Perform specific mappings for slots since PoB uses
		// a different scheme than the item api
		switch slotName {
		case "Weapon":
			slotName = "Weapon 1"
		case "Weapon2":
			slotName = "Weapon 1 Swap"
		case "Offhand":
			slotName = "Weapon 2"
		case "Offhand2":
			slotName = "Weapon 2 Swap"
		case "Helm":
			slotName = "Helmet"
		case "BodyArmour":
			slotName = "Body Armour"
		case "Flask":
			// Flask indexing starts at 1
			slotName = fmt.Sprintf("Flask %d", i.X+1)
		case "Ring":
			slotName = "Ring 1"
		case "Ring2":
			slotName = "Ring 2"
		}

		outSlot := Slot{
			Active: "true",
			Name:   slotName,
			ItemID: idString,
		}
		slotOut = append(slotOut, outSlot)
	}

	// ItemSet requires ALL items be present
	// or bad things happen.
	itemSetSlots := make([]Slot, 0, len(slotOut)+3)
	for _, s := range slotOut {
		itemSetSlots = append(itemSetSlots, s)
	}
	for _, req := range requiredSlots {
		found := false
		for _, s := range slotOut {
			if s.Name == req {
				found = true
				break
			}
		}

		if found {
			continue
		}
		itemSetSlots = append(itemSetSlots, Slot{
			Active: "",
			Name:   req,
			// 0 is a reserved identifier for 'not present'
			// Why? pob is written in lua, it starts its indexing at 1 D:
			ItemID: "0",
		})
	}

	activeSlot := "1"
	out := ItemsUnion{
		ActiveItemSet:      activeSlot,
		UseSecondWeaponSet: "false",

		Item: itemOut,
		Slot: slotOut,
		ItemSet: ItemSet{
			ID:   activeSlot,
			Slot: itemSetSlots,
		},
	}

	return out
}

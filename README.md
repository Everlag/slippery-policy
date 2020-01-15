# slippery-policy

Slippery-policy provides enforcement for the gucci-hobo rules of a Path of Exile private league:

- No non-unique, non-flask items equipped(past level 2)
- No private profiles

This outputs to a CSV file; the default output location `policy_failures.csv`. If that file is already is present, it is appended to rather than overwritten.

Best-effort deduplication of character policy failures past the first is performed. Across restarts of the tool, it may output duplicate entries; this can be cleaned up in post-processing.

Rate-limiting headers from GGG are respected.

Additional flags can be found in the cli interface using `./watch --help`

### Sample Output

This is a subset of the output from running against the `Slippery Hobo League (PL5357)` ladder. (This league completed prior to the tool being written)

If the reason for the line is `NonUniqueItemPresent`, additional information is filled out to provide context.

```
reason,itemName,itemLevel,itemSlot,characterName,characterLevel,accountName,when
NonUniqueItemPresent,Apocalypse Pelt Full Chainmail,73,BodyArmour,iakrana_hobo,92,iakrana,2020-01-15T04:43:27Z
NonUniqueItemPresent,Death Slicer Imperial Claw,68,Weapon,Musty_Hobo_Scion,90,BruceeFRost,2020-01-15T04:44:09Z
NonUniqueItemPresent,Primordial Staff,69,Weapon,BarniHobbo,89,Shadowtitan,2020-01-15T04:44:15Z
PrivateProfile,,0,,Gimmeluck,0,frankenmolar,2020-01-15T04:44:49Z
NonUniqueItemPresent,Highborn Bow,74,Weapon2,Meleeistherealchallenge,87,Tomberry,2020-01-15T04:44:49Z
```

### Building

This depends on a [go compiler](https://golang.org/doc/install).

The watch binary can be built from `cmd/watch` using `go build`.

A Makefile is present that provides significantly more ergonomics over running go commands manually.

### Etc

This is factored out of poe-diff, a project that provided historical progress tracking for Path of Exile characters. It's output [looked like this](https://gfycat.com/ScornfulMajorBubblefish).

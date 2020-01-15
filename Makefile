WATCH_DIR=cmd/watch

WORK_DIR=sandbox

FIXTURE_DIR=fixtures
FIXTURE_ACCOUNT ?= Everlag
FIXTURE_CHAR ?= SleeperSpectreBoi

LADDER_BASE=http://api.pathofexile.com/ladders
LADDER=Slippery%20Hobo%20League%20(PL5357)

GET_ITEMS_ENDPOINT=https://www.pathofexile.com/character-window/get-items
GET_PASSIVES_ENDPOINT=https://www.pathofexile.com/character-window/get-passive-skills

depends:
	go get -u github.com/gobuffalo/packr/...

.PHONY: test
test:
	@cd $(FIXTURE_DIR) && packr
	@GOCACHE=off go test ./...

.PHONY: win-release
win-release:
	@cd $(WATCH_DIR) && \
		GOOS=windows GOARCH=amd64 go build -o slippery-policy.exe

.PHONY: build-watch
build-watch:
	@cd $(WATCH_DIR) && \
		go build

.PHONY: run-watch
run-watch: build-watch
	@echo "executing in $(WORK_DIR)"
	@rm -r $(WORK_DIR) && \
	mkdir -p $(WORK_DIR) && \
	cp $(WATCH_DIR)/watch $(WORK_DIR)/watch && \
	cd $(WORK_DIR) && \
	./watch

fixtures/get-items.raw.json:
	@echo "fetching $(FIXTURE_ACCOUNT)/$(FIXTURE_CHAR) items from remote"
	@curl -s -S -X POST \
		-F 'accountName=$(FIXTURE_ACCOUNT)' \
		-F "character=$(FIXTURE_CHAR)" \
		$(GET_ITEMS_ENDPOINT) > $@
	@echo "checking response"
	@! cat $@ | grep -s "Resource not found"
	@echo "successful item fetch"

fixtures/get-items.json: fixtures/get-items.raw.json
	@echo "prettifying fixtures/get-items.raw.json"
	@cat fixtures/get-items.raw.json | jq . > $@
	@echo "outputting to fixtures/get-items.json"

fixtures/get-passive-skills.raw.json:
	@echo "fetching $(FIXTURE_ACCOUNT)/$(FIXTURE_CHAR) passives from remote"
	@curl -s -S -X GET -G \
		--data-urlencode "accountName=$(FIXTURE_ACCOUNT)" \
		--data-urlencode "character=$(FIXTURE_CHAR)" \
		$(GET_PASSIVES_ENDPOINT) > $@
	@echo "checking response"
	@! cat $@ | grep -s "Resource not found"
	@echo "successful passives fetch"

fixtures/get-passive-skills.json: fixtures/get-passive-skills.raw.json
	@echo "get-passive-skills.raw.json"
	@cat fixtures/get-passive-skills.raw.json | jq . > $@
	@echo "outputting to $@"

fixtures/get-ladder.raw.json:
	@echo "fetching $(LADDER) ladder from remote"
	curl -s -S -X GET -G \
		'$(LADDER_BASE)/$(LADDER)?offset=0&limit=20' > $@
	@echo "checking response"
	@! cat $@ | grep -s "Resource not found"
	@echo "successful ladder fetch"

fixtures/get-ladder.json: fixtures/get-ladder.raw.json
	@echo "get-ladder.json"
	@cat fixtures/get-ladder.raw.json | jq . > $@
	@echo "outputting to $@"

fixtures-dir:
	mkdir -p $(FIXTURE_DIR)

.PHONY: fixtures
fixtures: fixtures-dir | fixtures/get-ladder.json fixtures/get-items.json fixtures/get-passive-skills.json
	@echo "populated $(FIXTURE_DIR)"
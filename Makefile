##
# Binaries
##

NODE_BIN := node_modules/.bin

DEPCHECK := $(NODE_BIN)/dependency-check
ESLINT := $(NODE_BIN)/eslint

##
# Files
##

BINS := $(shell find bin -type f)
SRCS := $(shell find lib -type f -name "*.js")
ALL_FILES := $(BINS) $(SRCS)

##
# Program flags
##

# Ignore any modules depcheck shouldn't care about, e.g. binaries
DEPCHECK_IGNORE_MODULES ?= \
	@segment/eslint-config \
	dependency-check \
	eslint \
	istanbul \
	mocha \
	mocha-circleci-reporter \
	eslint-plugin-require-path-exists

# Arguments for `check-dependencies`
DEPCHECK_FLAGS ?= \
	$(addprefix --ignore-module , $(DEPCHECK_IGNORE_MODULES))

##
# Tasks
##

# Installs node dependencies.
node_modules: package.json $(wildcard node_modules/*/package.json)
	@npm install
	@touch node_modules

# Installs dependencies.
install: node_modules
.PHONY: install

# Removes temporary files.
clean:
	rm -rf $(COVERAGE_DIR)
.PHONY: clean

# Removes temporary and vendor files.
distclean: clean
	rm -rf node_modules
.PHONY: distclean

# Runs linter and prints errors.
lint: install
	@$(ESLINT) $(ALL_FILES)
.PHONY: lint

# Runs linter, fixing as many errors as possible and printing remaining errors.
fmt: install
	@$(ESLINT) --fix $(ALL_FILES)
.PHONY: fmt

# Checks for unused or missing dependencies.
check-dependencies: install
	@$(DEPCHECK) --missing $(DEPCHECK_FLAGS) ./package.json
	@$(DEPCHECK) --unused $(DEPCHECK_FLAGS) ./package.json

# Runs all test tasks.
test: lint check-dependencies
.PHONY: test
.DEFAULT_GOAL = test

TEST_FLAGS ?= \
	-v \
	-cover

test:
	@go test $(TEST_FLAGS) ./...
.PHONY: test

test-ci: TEST_FLAGS += -p 1 -bench=. -benchmem
test-ci: test
.PHONY: test-ci

SHELL := /bin/bash

# Recursive wildcard function
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

# Generate the test binary name from the directory. Used in combination with
# .SECONDEXPANSION so that the pattern match from the target can be used more
# than once to generate the prerequisite.
testBinFromDir=$(addprefix $(addsuffix /,$(1)), $(addsuffix .test, $(notdir $(1))))

# Determine the list of go packages to be tested based on which have test files.
# Test targets are of the form `pkgdir/pkgname.test`
test_files := $(call rwildcard,,*_test.go)
pkgdirs := $(filter-out vendor/%, $(sort $(dir $(test_files))))
pkgs := $(notdir $(patsubst %/,%,$(pkgdirs)))
testBins := $(join $(pkgdirs), $(addsuffix .test,$(pkgs)))
testOutputs := $(addsuffix test.out,$(pkgdirs))

.PHONY: godocdown
godocdown:
	find -type f -name \*.go -not -path "./vendor/*" -execdir godocdown -template $(CURDIR)/.godocdown.template -o README.md \;

.PHONY: lint-required
lint-required:
	gometalinter @gometalinter.required.flags

.PHONY: lint-optional
lint-optional:
	gometalinter @gometalinter.optional.flags

# Suppress Make output. The relevant test output will be collected and sent to
# stdout
.SILENT:

# Test entry point to run all go package tests.
.PHONY: test
test: $(testOutputs)

# Run a package's test suite in a container and collect the output.
# Use a custom tmp directory so zpool commands work, which require paths
# relative to the host. Mount /tmp for that custom tmp directory to work.
# Mount the cgroup fs for systemd to work. Use SYS_ADMIN cap for zpool mounting.
# Add /dev/zfs device for zfs to work.
.SECONDEXPANSION:
$(testOutputs): %/test.out: $$(call testBinFromDir,%)
	flock /dev/stdout -c 'echo "RUN   $<"'
	./run-test.sh $<

# Build a package's test binaries. Done outside the container so it can be used
# with the base MistifyOS instead of the SDK. Always for a rebuild.
.SECONDARY: $(testBins)
$(testBins): %.test: FORCE
	echo BUILD $@
	cd $(dir $@) && flock -s /dev/stdout go test -c -i

FORCE:

clean:
	go clean ./...

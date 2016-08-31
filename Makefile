SHELL := /bin/bash

# Bash scripts known to be lint-clean
BASH_CLEAN=boot/*/* .travis/install-shellcheck.sh
# Bash scripts not yet lint-clean
BASH_DIRTY=.travis/deps.sh build/*.sh build/scripts/cerana-functions.sh build/scripts/start-vm build/scripts/vm-network

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

errcheckMatch := ^([^:]+):([[:digit:]]+):([[:digit:]]+)\t(.*)
errcheckReplace := \1:\2:\3:warning: error return value not checked (\4) (errcheck)

.PHONY: lint-required
lint-required:
	gometalinter @gometalinter.required.flags
	# errcheck has been separated here for efficiency.
	# It has to load all of the files to analyze APIs. When given multiple pkg
	# paths, it takes advantage of caching after loads. Using gometalinter runs
	# it separately for each pkg, losing the caching speed benefits.
	set -o pipefail; \
		errcheck $$(go list ./... | grep -v /vendor/) 2>&1 | \
		sed -r 's|$(errcheckMatch)|$(errcheckReplace)|'

.PHONY: lint-optional
lint-optional:
	gometalinter @gometalinter.optional.flags

.PHONY: bash-lint-required
bash-lint-required:
	shfmt -i 4 -l $(BASH_CLEAN) | diff /dev/null -
	shellcheck $(BASH_CLEAN)

.PHONY: bash-lint-optional
bash-lint-optional:
	shfmt -i 4 -l $(BASH_DIRTY) | diff /dev/null -
	shellcheck $(BASH_DIRTY)

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
	echo "RUN   $<"
	./run-test.sh $<

# Build a package's test binaries. Done outside the container so it can be used
# with the base MistifyOS instead of the SDK. Always for a rebuild.
.SECONDARY: $(testBins)
$(testBins): %.test: FORCE
	echo BUILD $@
	cd $(dir $@) && go test -c -i

FORCE:

clean:
	go clean ./...

.PHONY: shfmt
shfmt:
	shfmt -i 4 -w $(BASH_CLEAN) $(BASH_DIRTY)

SHELL := /bin/bash
IMAGE := mistifyio/mistify-os:zfs-stable-api

# Recursive wildcard function
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

# Generate the test binary name from the directory. Used in combination with
# .SECONDEXPANSION so that the pattern match from the target can be used more
# than once to generate the prerequisite.
testBinFromDir=$(addprefix $(addsuffix /,$(1)), $(addsuffix .test, $(notdir $(1))))

# Determine the list of go packages to be tested based on which have test files.
# Test targets are of the form `pkgdir/pkgname.test`
test_files := $(call rwildcard,,*_test.go)
pkgdirs := $(sort $(dir $(test_files)))
pkgs := $(notdir $(patsubst %/,%,$(pkgdirs)))
testBins := $(join $(pkgdirs), $(addsuffix .test,$(pkgs)))
testOutputs := $(addsuffix test.out,$(pkgdirs))

.PHONY: godocdown
godocdown:
	find -type f -name \*.go -execdir godocdown -template $(CURDIR)/.godocdown.template -o README.md \;

.PHONY: lint-required
lint-required:
	cat gometalinter.required.flags <(printf "\n") <(go list ./... | sed "s|github.com/cerana/cerana/||" | grep -v -E "^(zfs|zfs/nv|cmd/zfs|cmd/nvprint)$$") | gometalinter @/dev/stdin

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
	tmpdir=/tmp/$(notdir $<); \
	rm -rf $$tmpdir && \
	mkdir $$tmpdir && \
	cid=$(shell docker run -dti -v "$(CURDIR):/mistify:ro" -v /tmp:/tmp -v /sys/fs/cgroup:/sys/fs/cgroup:ro --cap-add=SYS_ADMIN --device /dev/zfs:/dev/zfs --name $(notdir $<) -e "TMPDIR=$(tmpdir)" $(IMAGE)) && \
	test -n $(cid) && \
	sleep .25 && \
	docker exec $$cid sh -c "cd /mistify; cd $(@D); ./$(notdir $<) -test.v" &> $@; \
	ret=$$?; \
	docker kill $$cid  &>/dev/null && \
	docker rm -v $$cid &>/dev/null && \
	flock /dev/stdout -c 'echo "+++ $< +++"; cat $@'; \
	rm -rf $$tmpdir; \
	exit $$ret

# Build a package's test binaries. Done outside the container so it can be used
# with the base MistifyOS instead of the SDK. Always for a rebuild.
.SECONDARY: $(testBins)
$(testBins): %.test: FORCE
	echo BUILD $@
	cd $(dir $@) && flock -s /dev/stdout go test -c

FORCE:

clean:
	go clean ./...
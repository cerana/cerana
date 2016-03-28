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
pkgdirs := $(sort $(dir $(test_files)))
pkgs := $(notdir $(patsubst %/,%,$(pkgdirs)))
testBins := $(join $(pkgdirs), $(addsuffix .test,$(pkgs)))
testOutputs := $(addsuffix test.out,$(pkgdirs))

# Suppress Make output. The relevant test output will be collected and sent to
# stdout
.SILENT:

# Test entry point to run all go package tests.
.PHONY: test
test: $(testOutputs)

# Force the test targets to run. For some reason, just using phony doesn't work.
FORCE:

# Run a package's test suite in a container and collect the output.
.SECONDEXPANSION:
.PHONY: $(testOutputs)
$(testOutputs): %/test.out: $$(call testBinFromDir,%) FORCE
	flock /dev/stdout -c 'echo "RUN   $<"'
	cid=$(shell docker run -dti -v "$(CURDIR):/mistify:ro" -v /sys/fs/cgroup:/sys/fs/cgroup:ro --name $(notdir $<) mistifyio/mistify-os) && \
	test -n $(cid) && \
	sleep .25 && \
	docker exec $$cid sh -c "cd /mistify; cd $(@D); ./$(notdir $<) -test.v" &> $@; \
	ret=$$?; \
	docker kill $$cid  &>/dev/null && \
	docker rm -v $$cid &>/dev/null && \
	flock /dev/stdout -c 'echo "+++ $< +++"; cat $@'; \
	exit $$ret

# Build a package's test binaries. Done outside the container so it can be used
# with the baes MistifyOS instead of the SDK.
.SECONDARY: $(testBins)
$(testBins): %.test:
	echo BUILD $@
	cd $(dir $@) && flock -s /dev/stdout go test -c

clean:
	go clean ./...
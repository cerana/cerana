rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

SHELL := /bin/bash

test_files := $(call rwildcard,,*_test.go)
pkgdirs := $(sort $(dir $(test_files)))
pkgs := $(notdir $(patsubst %/,%,$(pkgdirs)))
testBins := $(addsuffix .test,$(pkgs))
tests := $(join $(pkgdirs), $(testBins))

.PHONY: test
test: $(addsuffix .run.out,$(tests))

.PHONY: %.test.run.out
%.test.run.out : %.test.run
	flock /dev/stdout -c 'cat $@'

.PHONY: %.test.run
%.test.run: %.test %
	flock /dev/stdout -c 'echo "RUN   $<"'
	cid=$(shell docker run -dti -v "$(CURDIR):/mistify:ro" -v /sys/fs/cgroup:/sys/fs/cgroup:ro --name $(notdir $<) mistifyio/mistify-os) && \
	test -n $(cid) && \
	sleep .25 && \
	docker exec $$cid sh -c "cd /mistify; cd $(@D); ./$(notdir $<) -test.v" &> $@.out; \
	ret=$$?; \
	docker kill $$cid  &>/dev/null && \
	docker rm -v $$cid &>/dev/null && \
	flock /dev/stdout -c 'echo "+++ $< +++"; cat $@.out'; \
	exit $$ret

.SECONDARY: $(tests)
%.test:
	echo BUILD $@
	cd $(dir $@) && flock -s /dev/stdout go test -c
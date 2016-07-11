DESTDIR=/does/not/exist

.PHONY: install install-scripts install-bins lint reformat

install: install-scripts install-bins

install-scripts: scripts/*
	install -D -m 755 -t $(DESTDIR)/scripts $?

install-bins: bin/*
	install -D -m 755 -t $(DESTDIR)/bin $?

reformat:
	shfmt -i 4 -w scripts/* bin/*

lint:
	shfmt -i 4 -l scripts/* bin/* | diff /dev/null -
	shellcheck scripts/* bin/*

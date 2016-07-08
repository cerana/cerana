DESTDIR=/does/not/exist

.PHONY: install install-scripts install-bins lint

install: install-scripts install-bins

install-scripts: scripts/*
	install -D -m 755 -t $(DESTDIR)/scripts $?

install-bins: bin/*
	install -D -m 755 -t $(DESTDIR)/bin $?

lint:
	shfmt -i 4 -w scripts/*
	shellcheck scripts/*

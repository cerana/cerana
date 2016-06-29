DESTDIR=/does/not/exist

.PHONY: install install-scripts install-bins lint

install: install-scripts install-bins

install-scripts: scripts/*
	mkdir -p $(DESTDIR)/scripts/
	cp $? $(DESTDIR)/scripts/
	chmod +x $(DESTDIR)/scripts/*

install-bins: bin/*
	mkdir -p $(DESTDIR)/bin/
	cp $? $(DESTDIR)/bin/
	chmod +x $(DESTDIR)/bin/*

lint:
	shfmt -i 4 -w scripts/*
	shellcheck scripts/*

DESTDIR=/does/not/exist

.PHONY: install install-scripts install-bins

install: install-scripts install-bins

install-scripts: scripts/*
	mkdir -p $(DESTDIR)/scripts/
	cp $? $(DESTDIR)/scripts/
	chmod +x $(DESTDIR)/scripts/*

install-bins: bin/*
	mkdir -p $(DESTDIR)/bin/
	cp $? $(DESTDIR)/bin/
	chmod +x $(DESTDIR)/bin/*

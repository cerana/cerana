DESTDIR=/does/not/exist

.PHONY: install install-scripts

install: install-scripts

install-scripts: scripts/*
	mkdir -p $(DESTDIR)/scripts/
	cp $? $(DESTDIR)/scripts/
	chmod +x $(DESTDIR)/scripts/*

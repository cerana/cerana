EXTEN=svg

view: boot.$(EXTEN)
	open $<

%.$(EXTEN): %.dot
	dot -T$(EXTEN) $< > $@

.PHONY: clean view

clean:
	rm -f *.svg

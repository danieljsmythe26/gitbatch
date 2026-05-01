.PHONY: install update install-wrapper

install:
	./scripts/gb install

update:
	./scripts/gb update

install-wrapper:
	install -m 0755 ./scripts/gb $(HOME)/.local/bin/gb

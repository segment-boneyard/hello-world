
build: install

install: node_modules

node_modules: package.json
	@-rm -r node_modules
	@npm install

clean:
	@rm -Rf node_modules

.PHONY: install build clean

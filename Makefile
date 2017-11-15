# generate version number
version=$(shell git describe --tags --long --always|sed 's/^v//')

all: dep
	go fmt


dep:
	gom install

clean:
	rm -rf _vendor
version:
	@echo $(version)

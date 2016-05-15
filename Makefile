APPC_SPEC_VERSION=v0.7.4

CC = clang
GOPATH = $(.CURDIR)/build
GO15VENDOREXPERIMENT = 1
.export CC GOPATH GO15VENDOREXPERIMENT
gopkg = github.com/3ofcoins/appc-metadata-client
.gopath = ${GOPATH}/src/${gopkg}

ac-mdc: ${.gopath} mdc.go
	go build -o ac-mdc${FLAVOUR:D.}${FLAVOUR}  ${gopkg}

test: ${.gopath} .PHONY
	go test ${gopkg}

.gopath: ${.gopath}
${.gopath}:
	mkdir -p "${.gopath:H}"
	rm -rf "${.gopath}"
	ln -sv ${.CURDIR} ${.gopath}

clean:
	rm -rf ac-mdc build

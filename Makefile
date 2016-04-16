APPC_SPEC_VERSION=v0.7.4

CC = clang
GOPATH = $(.CURDIR)/build
GO15VENDOREXPERIMENT = 1
.export CC GOPATH GO15VENDOREXPERIMENT
gopkg = github.com/3ofcoins/appc-metadata-client
.gopath = ${GOPATH}/src/${gopkg}

ac-mdc: ${.gopath} mdc.go
	go build -o ac-mdc  ${gopkg}

test: ${.gopath} .PHONY
	go test ${gopkg}

.gopath: ${.gopath}
${.gopath}:
	mkdir -p "${.gopath:H}"
	rm -rf "${.gopath}"
	ln -sv ${.CURDIR} ${.gopath}

vendor.refetch: .PHONY
	rm -rf vendor build
	${MAKE} ${.gopath}
	go get -v ${gopkg}
	mv -v ${.CURDIR}/build/src ${.CURDIR}/vendor
	rm -rfv vendor/github.com/3ofcoins
	cd ${.CURDIR}/vendor/github.com/appc/spec && git checkout ${APPC_SPEC_VERSION}
	set -e ; \
	    cd ${.CURDIR}/vendor ; \
	     for d in github.com/*/* ; do \
	         if test -L $$d ; then \
	             continue ; \
	         fi ; \
	         echo "$$d $$(cd $$d; git log -n 1 --oneline --decorate)" >> $(.CURDIR)/vendor/manifest.txt ; \
	         rm -rf $$d/.git ; \
             done

clean:
	rm -rf ac-mdc build

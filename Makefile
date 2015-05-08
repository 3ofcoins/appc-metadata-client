APPC_SPEC_VERSION=v0.5.1

CC=clang
GOPATH=$(.CURDIR)/vendor
.export CC GOPATH

mdc: mdc.go
	go build -o mdc

vendor.refetch: .PHONY
	rm -rf vendor
	go get -d
	cd ${.CURDIR}/vendor/src/github.com/appc/spec && git checkout ${APPC_SPEC_VERSION}
	set -e ; \
	    cd ${.CURDIR}/vendor/src ; \
	    for d in github.com/*/* ; do \
	        if test -L $$d ; then \
	            continue ; \
	        fi ; \
	        echo "$$d $$(cd $$d; git log -n 1 --oneline --decorate)" >> $(.CURDIR)/vendor/manifest.txt ; \
	        rm -rf $$d/.git ; \
            done

clean:
	rm -rf mdc

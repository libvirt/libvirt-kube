
COMMANDS = virtkubecri virtkubevmshim virtkubenodeinfo virtkubeimagerepo virtkubevmangel

GOPATH = $(shell echo $$GOPATH)

BINARIES = $(COMMANDS:%=$(GOPATH)/bin/%)

SRC = $(shell find pkg -name '*.go')

all: $(BINARIES)

# We don't build images by default, since it is valid to
# run the binaries direct from host context, and building
# docker images massively slows down rebuild time during
# dev.
images:
	(cd images && ./build.sh)

clean:
	for c in $(COMMANDS); do rm -f $(GOPATH)/bin/$$c ; done

$(BINARIES): .vendor.status $(SRC)

.vendor.status: glide.yaml
	if test -d vendor; then \
		glide update --strip-vendor; \
	else \
		glide install --strip-vendor; \
	fi && touch $@


$(GOPATH)/bin/%: cmd/%
	go install libvirt.org/libvirt-kube/$<

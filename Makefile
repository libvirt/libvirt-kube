
COMMANDS = virtkubecri virtkubevmshim virtkubenodeinfo virtkubeimagerepo virtkubevmangel

BINARIES = $(COMMANDS:%=build/%)

SRC = $(shell find pkg -name '*.go')

TEST_DIRS = libvirt/config

all: $(BINARIES)

check:
	go test $(TEST_DIRS:%=libvirt.org/libvirt-kube/pkg/%)

$(BINARIES): .vendor.status $(SRC)

.vendor.status: glide.yaml
	if test -d vendor; then \
		glide update --strip-vendor; \
	else \
		glide install --strip-vendor; \
	fi && touch $@


build/%: cmd/%
	go build -o $@ ./$<

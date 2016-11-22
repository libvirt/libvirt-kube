
COMMANDS = libvirt-kubelet

BINARIES = $(COMMANDS:%=build/%)

TEST_DIRS = libvirt/config

all: $(BINARIES)

check:
	go test $(TEST_DIRS:%=libvirt.org/libvirt-kubelet/pkg/%)

$(BINARIES): .vendor.status

.vendor.status: glide.yaml
	if test -d vendor; then \
		glide update --strip-vendor --quick; \
	else \
		glide install --strip-vendor; \
	fi && touch $@


build/libvirt-kubelet: cmd/libvirt-kubelet
	go build -o $@ ./$<

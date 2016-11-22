
COMMANDS = libvirt-kubelet

BINARIES = $(COMMANDS:%=build/%)

all: $(BINARIES)

$(BINARIES): .vendor.status

.vendor.status: glide.yaml
	if test -d vendor; then \
		glide update --strip-vendor --quick; \
	else \
		glide install --strip-vendor; \
	fi && touch $@


build/libvirt-kubelet: cmd/libvirt-kubelet
	go build -o $@ ./$<

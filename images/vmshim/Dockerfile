# We need libvirt 3.0.0, hence currently rawhide, not F25
FROM fedora:rawhide
MAINTAINER http://libvirt.org

RUN dnf -y install \
	libvirt-client \
	&& dnf clean all

VOLUME /run/virtkubevmshim
VOLUME /srv/libvirt/run

# The entrypoint.sh script runs before services start up to ensure that
# critical directories and permissions are correct.
COPY virtkubevmshim /usr/local/bin/virtkubevmshim

CMD ["/usr/local/bin/virtkubevmshim"]

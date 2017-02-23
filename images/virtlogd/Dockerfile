FROM fedora:25
MAINTAINER http://libvirt.org

# USING THIS CONTAINER
#
# # docker run --privileged -d \
# 	--net=host \
# 	-v /var/log/libvirt:/var/log/libvirt:rw \
# 	-v /run/libvirt:/run/libvirt:rw \
# 	libvirtkubevirtlogd

VOLUME /var/log/libvirt
VOLUME /run/libvirt

RUN dnf -y install \
	libvirt-daemon \
	&& dnf clean all

CMD ["/usr/sbin/virtlogd"]

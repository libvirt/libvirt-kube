# We need libvirt 3.0.0, hence currently rawhide, not F25
FROM fedora:rawhide
MAINTAINER http://libvirt.org


# USING THIS CONTAINER
#
# $ docker run --privileged -d \
# 	--net=host \
# 	-v /srv/libvirt:/srv/libvirt:rw \
# 	libvirtkubevirtlogd
#
# NB, this will provide an isolated virtlogd instance
# that is completely separate from any virtlogd on
# the host OS.
#
# If you wish to instead replace the host virtlogd,
# then setup bind mounts in /srv/libvirt to point to
# the various libvirt dirs eg
#
#   mkdir /srv/libvirt/{log,run}
#   mount --bind /var/log/libvirt /srv/libvirt/log
#   mount --bind /run/libvirt /srv/libvirt/run

RUN dnf -y install \
	libvirt-daemon \
	&& dnf clean all

VOLUME /srv/libvirt

# The entrypoint.sh script runs before services start up to ensure that
# critical directories and permissions are correct.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/sbin/virtlogd"]

# We need libvirt 3.0.0, hence currently rawhide, not F25
FROM fedora:rawhide
MAINTAINER http://libvirt.org

# USING THIS CONTAINER
#
# $ docker run --privileged -d \
# 	--net=host \
#	--device /dev/kvm:/dev/kvm \
# 	-v /srv/libvirt:/srv/libvirt:rw \
# 	-v /run/dbus:/run/dbus:rw \
# 	-v /sys/fs/cgroup:/sys/fs/cgroup:rw \
# 	libvirtkubevirtd
#
# NB, this will provide an isolated libvirtd instance
# that is completely separate from any libvirtd on
# the host OS.
#
# If you wish to instead replace the host libvirtd,
# then setup bind mounts in /srv/libvirt to point to
# the various libvirt dirs eg
#
#   mkdir /srv/libvirt/{etc,log,lib,run}
#   mkdir /srv/libvirt/etc/{qemu,lxc,secrets,storage,nwfilter}
#   mount --bind /etc/libvirt/qemu /srv/libvirt/etc/qemu
#   mount --bind /etc/libvirt/lxc /srv/libvirt/etc/lxc
#   mount --bind /etc/libvirt/secrets /srv/libvirt/etc/secrets
#   mount --bind /etc/libvirt/storage /srv/libvirt/etc/storage
#   mount --bind /etc/libvirt/nwfilter /srv/libvirt/etc/nwfilter
#   mount --bind /var/log/libvirt /srv/libvirt/log
#   mount --bind /var/lib/libvirt /srv/libvirt/lib
#   mount --bind /run/libvirt /srv/libvirt/run

RUN dnf -y install \
	libvirt-daemon-qemu \
	libvirt-daemon-kvm \
	libvirt-client \
	augeas \
	&& dnf clean all

RUN augtool -s set /files/etc/libvirt/libvirtd.conf/ca_file /etc/pki/libvirt/cacert.pem && \
    augtool -s set /files/etc/libvirt/libvirtd.conf/key_file /etc/pki/libvirt/serverkey.pem && \
    augtool -s set /files/etc/libvirt/libvirtd.conf/cert_file /etc/pki/libvirt/servercert.pem

VOLUME /srv/libvirt
VOLUME /sys/fs/cgroup
VOLUME /run/dbus

# The default configuration for this image permits unauthenticated tcp
# connections to libvirt on this port.  Do not publish this port without
# providing an alternative configuration.
EXPOSE 16509

# The entrypoint.sh script runs before services start up to ensure that
# critical directories and permissions are correct.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/sbin/libvirtd", "--listen"]

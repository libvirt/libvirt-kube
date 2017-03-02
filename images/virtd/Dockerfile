# We need libvirt 3.0.0, hence currently rawhide, not F25
FROM fedora:rawhide
MAINTAINER http://libvirt.org

# USING THIS CONTAINER
#
# # docker run --privileged -d \
# 	--net=host \
#	--device /dev/kvm:/dev/kvm \
# 	-v /var/log/libvirt:/var/log/libvirt:rw \
# 	-v /var/lib/libvirt:/var/lib/libvirt:rw \
# 	-v /run/libvirt:/run/libvirt:rw \
# 	-v /sys/fs/cgroup:/sys/fs/cgroup:rw \
# 	libvirtkubevirtd

VOLUME /var/lib/libvirt
VOLUME /var/log/libvirt
VOLUME /run/libvirt
VOLUME /sys/fs/cgroup

# The default configuration for this image permits unauthenticated tcp
# connections to libvirt on this port.  Do not publish this port without
# providing an alternative configuration.
EXPOSE 16509

RUN dnf -y install \
	libvirt-daemon-qemu \
	libvirt-daemon-kvm \
	libvirt-client \
	augeas \
	&& dnf clean all

RUN augtool -s set /files/etc/libvirt/libvirtd.conf/ca_file /etc/pki/libvirt/cacert.pem && \
    augtool -s set /files/etc/libvirt/libvirtd.conf/key_file /etc/pki/libvirt/serverkey.pem && \
    augtool -s set /files/etc/libvirt/libvirtd.conf/cert_file /etc/pki/libvirt/servercert.pem

# The entrypoint.sh script runs before services start up to ensure that
# critical directories and permissions are correct.
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/sbin/libvirtd", "--listen"]

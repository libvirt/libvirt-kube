# We need libvirt 3.0.0, hence currently rawhide, not F25
FROM fedora:rawhide
MAINTAINER http://libvirt.org

VOLUME /run/virtkubevmshim

# The entrypoint.sh script runs before services start up to ensure that
# critical directories and permissions are correct.
COPY virtkubevmangel /usr/local/bin/virtkubevmangel

CMD ["/usr/local/bin/virtkubevmangel"]

FROM scratch

ADD rootfs /

RUN apk update \
	&& apk add openrc openssh sudo util-linux curl bash nerdctl iptables \
	&& ssh-keygen -A \
	&& ln -s agetty /etc/init.d/agetty.ttyS0 \
	&& echo ttyS0 > /etc/securetty \
	&& rc-update add agetty.ttyS0 default \
	&& rc-update add devfs boot \
	&& rc-update add procfs boot \
	&& rc-update add sysfs boot \
    && rc-update add cgroups boot \
    && rc-update add networking sysinit \
	&& rc-update add local default \
	&& rc-update add containerd default

RUN mkdir -p /run/openrc \
	&& touch /run/openrc/softlevel \
	&& rc-update add sshd

# configure sshd
RUN sed -i 's/^#\(PermitRootLogin\) .*/\1 yes/' /etc/ssh/sshd_config && \
    sed -i 's/^#\(PasswordAuthentication\) .*/\1 no/' /etc/ssh/sshd_config

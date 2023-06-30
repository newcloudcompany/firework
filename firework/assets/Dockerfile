FROM scratch

ADD rootfs /

WORKDIR /firework

RUN chmod 1777 /tmp

RUN echo -n "deb http://deb.debian.org/debian/ bullseye main\ndeb http://security.debian.org/debian-security bullseye/updates main" > /etc/apt/sources.list

RUN export DEBIAN_FRONTEND=noninteractive && apt update && apt install -y \
    iproute2 \
    ca-certificates \
    curl \
    dnsutils \
    iputils-ping \
    procps \
    cpu-checker && \
    apt clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Generate guest SSH keys
# RUN ssh-keygen -A

# RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

# configure sshd
# RUN sed -i 's/^#\(PermitRootLogin\) .*/\1 yes/' /etc/ssh/sshd_config && \
#     sed -i 's/^#\(PasswordAuthentication\) .*/\1 no/' /etc/ssh/sshd_config

# COPY config config
# COPY kernel kernel
# COPY vm vm
# COPY macgen.py .
# COPY *.sh .
# COPY Dockerfile .
# COPY .gitignore .
COPY init /sbin/init

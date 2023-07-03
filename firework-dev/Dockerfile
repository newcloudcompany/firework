# syntax=docker/dockerfile:1.3-labs

FROM scratch

ADD debian-bullseye-rootfs /

WORKDIR /firework

# apt doesn't work properly without this somehow
RUN chmod 1777 /tmp

COPY <<EOF /etc/apt/sources.list
deb http://deb.debian.org/debian/ bullseye main
deb-src http://deb.debian.org/debian/ bullseye main

deb http://security.debian.org/debian-security bullseye-security main
deb-src http://security.debian.org/debian-security bullseye-security main

deb http://deb.debian.org/debian/ bullseye-updates main
deb-src http://deb.debian.org/debian/ bullseye-updates main
EOF

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

COPY init /sbin/init

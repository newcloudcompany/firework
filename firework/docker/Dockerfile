FROM debian

RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        iptables \
        iproute2 \
        git && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Download Firecracker
RUN curl -fL -o /root/firecracker-v1.3.3-x86_64.tgz \
    https://github.com/firecracker-microvm/firecracker/releases/download/v1.3.3/firecracker-v1.3.3-x86_64.tgz

# Unarchive
RUN tar -xvf /root/firecracker-v1.3.3-x86_64.tgz -C /root

# Move binary to /usr/bin
RUN mv /root/release-v1.3.3-x86_64/firecracker-v1.3.3-x86_64 /usr/bin/firecracker
RUN chmod +x /usr/bin/firecracker
#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

function install_containerd {
    curl -OL https://github.com/containerd/containerd/releases/download/v1.7.2/containerd-1.7.2-linux-amd64.tar.gz
    tar Cxzvf /usr/local containerd-1.7.2-linux-amd64.tar.gz
    ln -sf "/etc/systemd/system/containerd.service" "/etc/systemd/system/multi-user.target.wants/containerd.service"
}

function install_runc {
    curl -OL https://github.com/opencontainers/runc/releases/download/v1.1.7/runc.amd64
    chmod +x runc.amd64
    install -m 755 runc.amd64 /usr/local/sbin/runc
}

function install_cni_plugins {
    curl -OL https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz
    mkdir -p /opt/cni/bin
    tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.3.0.tgz
}

function install_kubeadm_prerequisities {
    install_containerd
    install_runc
    install_cni_plugins
}

function install_kubeadm {
    apt install -y gnupg
    curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /etc/apt/keyrings/kubernetes-archive-keyring.gpg
    echo "deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | tee /etc/apt/sources.list.d/kubernetes.list
    apt update
    apt install -y kubelet kubeadm kubectl
    apt-mark hold kubelet kubeadm kubectl
}

install_kubeadm_prerequisities
install_kubeadm
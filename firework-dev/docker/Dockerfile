# syntax=docker/dockerfile:1.3-labs

FROM scratch

ARG ROOTFS_ARCHIVE_PATH

ADD $ROOTFS_ARCHIVE_PATH /

RUN apt clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

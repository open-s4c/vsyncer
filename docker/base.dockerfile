################################################################################
# base image
################################################################################
ARG FROM_IMAGE=ubuntu:22.04
FROM ${FROM_IMAGE} as base

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     llvm \
     clang \
     libclang-dev \
     llvm-dev \
     git \
     libz-dev \
     ca-certificates \
 && rm -rf /var/lib/apt/lists/*


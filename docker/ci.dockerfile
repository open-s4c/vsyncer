ARG TAG=main
ARG REPO=ghcr.io/open-s4c/

FROM ${REPO}vsyncer:${TAG}

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    make \
    ninja-build \
    libc-dev \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

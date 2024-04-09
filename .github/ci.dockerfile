ARG VSYNCER_TAG=main
FROM ghcr.io/open-s4c/vsyncer:$VSYNCER_TAG

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    make \
    ninja-build \
    libc-dev \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

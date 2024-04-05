FROM ghcr.io/open-s4c/vsyncer:main

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    make \
    ninja-build \
    libc-dev \
    golang-go \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

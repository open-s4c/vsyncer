ARG VSYNCER_TAG
FROM ghcr.io/open-s4c/vsyncer:sha-${VSYNCER_TAG}

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    make \
    ninja-build \
    libc-dev \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

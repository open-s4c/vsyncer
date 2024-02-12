FROM ubuntu:22.04 as llvm

RUN apt-get update && apt-get install -y --no-install-recommends \
    llvm \
    clang \
    libclang-dev \
    llvm-dev \
    git \
    libz-dev \
    && rm -rf /var/lib/apt/lists/*

FROM ubuntu:22.04 as go_builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go \
    make \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

FROM llvm as genmc_builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    make \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Note: The install prefix in the builder must match the install location on the final image.
RUN cd /tmp \
    && git clone https://github.com/open-s4c/genmc.git genmc9 \
    && cd genmc9 \
    && git checkout "v0.9" \
    && autoreconf --install \
    && ./configure --prefix=/usr/share/genmc9 \
    && make install -j8 \
    && rm -rf /tmp/genmc9

RUN cd /tmp \
    && git clone https://github.com/open-s4c/genmc.git genmc10 \
    && cd genmc10 \
    && git checkout "v0.10.0" \
    && autoreconf --install \
    && ./configure --prefix=/usr/share/genmc10 \
    && make install -j8 \
    && rm -rf /tmp/genmc10

FROM go_builder as vsyncer_builder
ARG VSYNCER_TAG=main
RUN cd tmp \
    && git clone https://github.com/open-s4c/vsyncer.git \
    && cd vsyncer \
    && git checkout "${VSYNCER_TAG}" \
    && make \
    && ./vsyncer --help \
    && cp ./vsyncer /usr/bin/vsyncer \
    && rm -rf /tmp/vsyncer

FROM llvm as final

COPY --from=genmc_builder /usr/share/genmc9 /usr/share/genmc9
COPY --from=genmc_builder /usr/share/genmc10 /usr/share/genmc10
COPY --from=vsyncer_builder /usr/bin/vsyncer /usr/bin/vsyncer

ENV PATH="/usr/share/genmc9/bin:$PATH"

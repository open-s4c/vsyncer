ARG BASE_TAG=dev
ARG BASE_IMAGE=ghcr.io/open-s4c/vsyncer/base

################################################################################
# genmc_builder
################################################################################
FROM ${BASE_IMAGE}:${BASE_TAG} as genmc_builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     autoconf \
     automake \
     make \
 && rm -rf /var/lib/apt/lists/*

# Note: The install prefix in the builder must match the install location on
# the final image.

RUN cd /tmp \
 && git clone https://github.com/open-s4c/genmc.git genmc9 \
 && cd genmc9 \
 && git checkout "v0.9" \
 && autoreconf --install \
 && ./configure --prefix=/usr/share/genmc9 \
 && make install -j8

RUN cd /tmp \
 && git clone https://github.com/open-s4c/genmc.git genmc10 \
 && cd genmc10 \
 && git checkout "v0.10.1-a" \
 && autoreconf --install \
 && ./configure --prefix=/usr/share/genmc10 \
 && make install -j8


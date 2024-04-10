# This is a multi-stage dockerfile to build vsyncer and its dependencies
ARG TAG=main

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

################################################################################
# genmc_builder
################################################################################
FROM base as genmc_builder

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

################################################################################
# dat3m_builder
################################################################################
FROM base as dat3m_builder

RUN apt-get update  \
 && apt-get install -y --no-install-recommends \
     graphviz \
     maven \
     autoconf \
     automake  \
     openjdk-17-jdk \
     openjdk-17-jre \
 && rm -rf /var/lib/apt/lists/*

RUN cd /tmp \
 && git clone --branch "4.0.0" --depth 1 \
     https://github.com/hernanponcedeleon/dat3m.git

RUN cd /tmp/dat3m \
 && mvn clean install -DskipTests \
 && mkdir -p /usr/share/dat3m/dartagnan \
 && cp -R dartagnan/target /usr/share/dat3m/dartagnan \
 && cp -R include /usr/share/dat3m \
 && cp -R cat /usr/share/dat3m

################################################################################
# vsyncer_builder
################################################################################
FROM base as vsyncer_builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     golang-go \
     make \
     git \
 && rm -rf /var/lib/apt/lists/*

RUN cd /tmp \
 && git clone https://github.com/open-s4c/vsyncer.git \
 && cd vsyncer \
 && git checkout "$TAG"

RUN cd /tmp/vsyncer \
 && make build \
 && make install PREFIX=/usr \
 && make clean \
 && vsyncer --help

################################################################################
# vsyncer image
################################################################################
FROM base as final

# basic tools
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     less vim \
 && rm -rf /var/lib/apt/lists/*

# dat3m
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     openjdk-17-jre \
 && rm -rf /var/lib/apt/lists/*

COPY --from=dat3m_builder /usr/share/dat3m /usr/share/dat3m
RUN ln -s /usr/share/dat3m/dartagnan/target/libs/*.so /usr/lib/
ENV DAT3M_HOME=/usr/share/dat3m
ENV DAT3M_OUTPUT="/tmp/dat3m"
#ENV CFLAGS="-I$DAT3M_HOME/include"
#ENV OPTFLAGS="-mem2reg -sroa -early-cse -indvars -loop-unroll -fix-irreducible -loop-simplify -simplifycfg -gvn"

# genmc
COPY --from=genmc_builder /usr/share/genmc9 /usr/share/genmc9
COPY --from=genmc_builder /usr/share/genmc10 /usr/share/genmc10
ENV PATH="/usr/share/genmc9/bin:$PATH"

# vsyncer
COPY --from=vsyncer_builder /usr/bin/vsyncer /usr/bin/vsyncer

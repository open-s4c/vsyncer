# This is a multi-stage dockerfile to build vsyncer and its dependencies

ARG FROM_IMAGE=ubuntu:22.04

################################################################################
# builder image
################################################################################
FROM ${FROM_IMAGE} as builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     clang \
     libclang-dev \
     llvm \
     llvm-dev \
     git \
     libz-dev \
     ca-certificates \
 && rm -rf /var/lib/apt/lists/*

################################################################################
# genmc_builder
################################################################################
FROM builder as genmc_builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     autoconf \
     automake \
     make \
 && rm -rf /var/lib/apt/lists/*

# Note: The install prefix in the builder must match the install location on
# the final image.

RUN cd /tmp \
 && git clone --depth 1 --branch "v0.9" \
     https://github.com/open-s4c/genmc.git genmc9

RUN cd /tmp/genmc9 \
 && autoreconf --install \
 && ./configure --prefix=/usr/share/genmc9 \
 && make install -j8

RUN cd /tmp \
 && git clone --depth 1 --branch "v0.10.1-a" \
     https://github.com/open-s4c/genmc.git genmc10

RUN cd /tmp/genmc10 \
 && autoreconf --install \
 && ./configure --prefix=/usr/share/genmc10 \
 && make install -j8

################################################################################
# dat3m_builder
################################################################################
FROM builder as dat3m_builder

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
 && git clone \
     https://github.com/hernanponcedeleon/dat3m.git \
 && cd dat3m \
 && git checkout "c48cc9f2b726f0f99947dacfbd88c99b281ff173"

RUN cd /tmp/dat3m \
 && mvn clean install -DskipTests \
 && mkdir -p /usr/share/dat3m/dartagnan \
 && cp -R dartagnan/target /usr/share/dat3m/dartagnan \
 && cp -R include /usr/share/dat3m \
 && cp -R cat /usr/share/dat3m \
 && cp pom.xml /usr/share/dat3m/pom.xml

################################################################################
# vsyncer_builder
################################################################################
FROM builder as vsyncer_builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     golang-go \
     make \
     git \
 && rm -rf /var/lib/apt/lists/*

ARG VSYNCER_TAG=main
RUN cd /tmp \
 && git clone https://github.com/open-s4c/vsyncer.git \
 && cd vsyncer \
 && git checkout "$VSYNCER_TAG"

RUN cd /tmp/vsyncer \
 && make build \
 && make install PREFIX=/usr \
 && make clean \
 && vsyncer --help

################################################################################
# vsyncer image
################################################################################
FROM ${FROM_IMAGE} as final

# tools
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     clang \
     less \
     libclang-dev \
     llvm \
     llvm-dev \
     openjdk-17-jre \
     vim \
 && rm -rf /var/lib/apt/lists/*

# dat3m
COPY --from=dat3m_builder /usr/share/dat3m /usr/share/dat3m
COPY --from=dat3m_builder /usr/share/dat3m/pom.xml /usr/share/dat3m/pom.xml
RUN ln -s /usr/share/dat3m/dartagnan/target/libs/*.so /usr/lib/
ENV DAT3M_HOME=/usr/share/dat3m
ENV DAT3M_OUTPUT="/tmp/dat3m"

# genmc
COPY --from=genmc_builder /usr/share/genmc9 /usr/share/genmc9
COPY --from=genmc_builder /usr/share/genmc10 /usr/share/genmc10
ENV PATH="/usr/share/genmc9/bin:$PATH"

# vsyncer
COPY --from=vsyncer_builder /usr/bin/vsyncer /usr/bin/vsyncer
ENV VSYNCER_DOCKER=false

# done

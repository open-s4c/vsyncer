ARG VSYNCER_TAG=main
ARG TAG=main
ARG REPO=ghcr.io/open-s4c/

FROM ${REPO}vsyncer-base:${TAG} as base
FROM ${REPO}vsyncer-genmc:${TAG} as genmc
FROM ${REPO}vsyncer-dat3m:${TAG} as dat3m

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
 && ls -la \
 && rm -r vsyncer \
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
FROM base as final

# basic tools
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     less vim \
 && rm -rf /var/lib/apt/lists/*

# dat3m dependencies
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     openjdk-17-jre \
 && rm -rf /var/lib/apt/lists/*

COPY --from=dat3m /usr/share/dat3m /usr/share/dat3m
RUN ln -s /usr/share/dat3m/dartagnan/target/libs/*.so /usr/lib/
ENV DAT3M_HOME=/usr/share/dat3m
ENV DAT3M_OUTPUT="/tmp/dat3m"

# genmc
COPY --from=genmc /usr/share/genmc9 /usr/share/genmc9
COPY --from=genmc /usr/share/genmc10 /usr/share/genmc10
ENV PATH="/usr/share/genmc9/bin:$PATH"

# vsyncer
COPY --from=vsyncer_builder /usr/bin/vsyncer /usr/bin/vsyncer

ARG BASE_TAG=dev
ARG BASE_IMAGE=ghcr.io/open-s4c/vsyncer/base

################################################################################
# dat3m_builder
################################################################################
FROM ${BASE_IMAGE}:${BASE_TAG} as dat3m_builder

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


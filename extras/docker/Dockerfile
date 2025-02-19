FROM docker.io/ubuntu:noble-20240423

ENV DUCKDB_VERSION 1.2.0
ENV FPM_VERSION 1.15.1
ENV GO_VERSION 1.24.0
ENV NODE_VERSION 22.14.0
ENV UV_VERSION 0.6.0

ENV DEBIAN_FRONTEND noninteractive

RUN groupadd -g 5000 dev \
    && useradd -u 5000 -g 5000 -m -s /bin/bash dev

RUN apt update \
    && apt install -y \
        apt-transport-https \
        bindfs \
        binutils \
        curl \
        g++ \
        gcc \
        git \
        gpg \
        jq \
        less \
        make \
        nano \
        netcat-traditional \
        supervisor \
        rpm \
        ruby \
        ruby-dev \
        tar \
        telnet \
        unzip \
        xz-utils \
        wget \
    && apt clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -L -s https://go.dev/dl/go${GO_VERSION}.linux-$(dpkg --print-architecture).tar.gz | tar xvfz - -C /usr/local \
    && echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/dev/.bashrc

RUN ARCHITECTURE=$(uname -m) \
    && case $ARCHITECTURE in \
            x86_64)  ARCHITECTURE='x64' ;; \
            aarch64) ARCHITECTURE='arm64' ;; \
            *)       echo "unknown architecture: $ARCHITECTURE" && exit 1 ;; \
        esac \
    && mkdir -p /usr/local/node \
    && curl -L -s https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${ARCHITECTURE}.tar.xz | tar xvfJ - -C /usr/local/node --strip-components 1 \
    && echo 'export PATH=$PATH:/usr/local/node/bin' >> /home/dev/.bashrc

RUN curl -L -s https://github.com/astral-sh/uv/releases/download/${UV_VERSION}/uv-$(uname -m)-unknown-linux-gnu.tar.gz | tar xvfz - -C /usr/local/bin --strip-components 1

RUN gem install fpm -v ${FPM_VERSION} \
    && gem cleanup

RUN curl -fsSL https://packagecloud.io/varnishcache/varnish60lts/gpgkey | \
        gpg --dearmor > /etc/apt/trusted.gpg.d/varnish-6.0.gpg \
    && echo 'deb [signed-by=/etc/apt/trusted.gpg.d/varnish-6.0.gpg] https://packagecloud.io/varnishcache/varnish60lts/ubuntu noble main' > /etc/apt/sources.list.d/varnish-6.0.list \
    && echo 'deb-src [signed-by=/etc/apt/trusted.gpg.d/varnish-6.0.gpg] https://packagecloud.io/varnishcache/varnish60lts/ubuntu noble main' >> /etc/apt/sources.list.d/varnish-6.0.list \
    && echo 'Package: varnish varnish-*' > /etc/apt/preferences.d/varnish \
    && echo 'Pin: release o=packagecloud.io/varnishcache/*' >> /etc/apt/preferences.d/varnish \
    && echo 'Pin-Priority: 1000' >> /etc/apt/preferences.d/varnish \
    && apt update \
    && apt install -y varnish \
    && apt clean \
    && rm -rf /var/lib/apt/lists/* \
    && usermod -a -G varnish dev

RUN curl -L -s https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/duckdb_cli-linux-$(dpkg --print-architecture).zip -o /tmp/duckdb_cli-linux.zip \
    && unzip /tmp/duckdb_cli-linux.zip -d /usr/local/bin \
    && rm /tmp/duckdb_cli-linux.zip

COPY ./supervisord.conf /etc/supervisor/conf.d/
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

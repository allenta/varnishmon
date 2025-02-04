FROM docker.io/ubuntu:noble-20240423

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

RUN curl -L -s https://go.dev/dl/go1.23.5.linux-$(dpkg --print-architecture).tar.gz | tar xvfz - -C /usr/local \
    && echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/dev/.bashrc

RUN ARCHITECTURE=$(uname -m) \
    && case $ARCHITECTURE in \
            x86_64)  ARCHITECTURE='x64' ;; \
            aarch64) ARCHITECTURE='arm64' ;; \
            *)       echo "unknown architecture: $ARCHITECTURE" && exit 1 ;; \
        esac \
    && mkdir -p /usr/local/node \
    && curl -L -s https://nodejs.org/dist/v22.13.1/node-v22.13.1-linux-${ARCHITECTURE}.tar.xz | tar xvfJ - -C /usr/local/node --strip-components 1 \
    && echo 'export PATH=$PATH:/usr/local/node/bin' >> /home/dev/.bashrc

RUN curl -L -s https://github.com/astral-sh/uv/releases/download/0.5.27/uv-$(uname -m)-unknown-linux-gnu.tar.gz | tar xvfz - -C /usr/local/bin --strip-components 1

RUN gem install fpm -v 1.15.1 \
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

RUN curl -L -s https://github.com/duckdb/duckdb/releases/download/v1.1.3/duckdb_cli-linux-$(dpkg --print-architecture).zip -o /tmp/duckdb_cli-linux.zip \
    && unzip /tmp/duckdb_cli-linux.zip -d /usr/local/bin \
    && rm /tmp/duckdb_cli-linux.zip

COPY ./supervisord.conf /etc/supervisor/conf.d/
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

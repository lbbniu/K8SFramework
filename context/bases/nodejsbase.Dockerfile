FROM node:lts-bullseye
COPY root /

ENV LANG en_US.utf8
ENV DEBIAN_FRONTEND=noninteractive

RUN rm -rf /bin/ls                                                                        \
# image debian:bullseye had "ls bug", we use busybox ls instead                           \
    && apt update                                                                         \
    && apt install                                                                        \
       ca-certificates openssl telnet curl wget default-mysql-client                      \
       iputils-ping vim tcpdump net-tools binutils procps tree                            \
       libssl-dev zlib1g-dev                                                              \
       tzdata localepurge busybox -y                                                      \
    && busybox --install                                                                  \
    && locale-gen en_US.utf8                                                              \
    && apt purge -y                                                                       \
    && apt clean all                                                                      \
    && rm -rf /var/lib/apt/lists/*                                                        \
    && rm -rf /var/cache/*.dat-old                                                        \
    && rm -rf /var/log/*.log /var/log/*/*.log                                             \
    && rm -rf /etc/localtime
# /etc/localtime will block container mount /etc/localtime from host

RUN mkdir -p /usr/local/app/tars                                                          \
    && npm install -g @tars/node-agent                                                    \
    && mv /usr/local/lib/node_modules/@tars/node-agent /usr/local/app/tars/               \
    && cd /usr/local/app/tars/node-agent                                                  \
    && npm install

ENV NODE_AGENT_BIN=/usr/local/app/tars/node-agent/bin/node-agent

RUN chmod +x /bin/entrypoint.sh
CMD ["/bin/entrypoint.sh"]
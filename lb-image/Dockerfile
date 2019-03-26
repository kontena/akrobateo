FROM alpine:3.8

RUN apk add -U --no-cache iptables sudo

RUN addgroup akrobateo && \
    adduser -G akrobateo -H -s /sbin/nologin -D akrobateo && \
    echo "akrobateo ALL=(root) NOPASSWD: /sbin/iptables" > /etc/sudoers.d/akrobateo

ADD entrypoint.sh /entrypoint.sh

USER akrobateo

CMD ["/entrypoint.sh"]
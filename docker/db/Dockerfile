FROM postgres:13-buster

MAINTAINER Alexander Harchenko <morontt@gmail.com>

ARG HOST_UID=1000

VOLUME ["/var/lib/postgresql/data"]

RUN usermod -u ${HOST_UID} postgres \
    && groupmod -g ${HOST_UID} postgres

EXPOSE 5432

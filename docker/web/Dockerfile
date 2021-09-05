FROM php:7.4-fpm

MAINTAINER Alexander Harchenko <morontt@gmail.com>

VOLUME ["/var/www/html"]

ARG HOST_UID=1000

ENV DEBIAN_FRONTEND noninteractive
ENV TZ "Europe/Kiev"

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
    && cp "$PHP_INI_DIR/php.ini-development" "$PHP_INI_DIR/php.ini" \
    && sed -i 's/;date.timezone =/date.timezone = Europe\/Kiev/' ${PHP_INI_DIR}/php.ini

RUN apt-get update \
    && apt-get install -y \
    libzip-dev libpq-dev \
    git \
    --no-install-recommends

RUN docker-php-ext-install -j$(nproc) zip pgsql pdo_pgsql \
    && yes | pecl install xdebug \
    && docker-php-ext-enable xdebug

# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Install dependencies
FROM ubuntu:latest AS build-env
RUN apt-get update
RUN apt-get install -y curl git unzip
RUN apt-get clean

# Clone the Flutter repo
RUN git clone https://github.com/flutter/flutter.git /usr/local/flutter

# Set Flutter path
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"

# Enable Flutter web
RUN flutter channel stable
RUN flutter upgrade
RUN flutter config --enable-web

# Run Flutter doctor
RUN flutter doctor -v

# Copy the app files to the container
COPY ../../../demo/app /usr/local/bin/app

# Set the working directory to the app files within the container
WORKDIR /usr/local/bin/app

# Get app dependenciesxw
RUN flutter pub get

# Build the app for the web
RUN flutter build web

# Create the run-time image
FROM nginx:stable-alpine3.17-slim
COPY --from=build-env /usr/local/bin/app/build/web /usr/share/nginx/html
COPY --from=build-env /usr/local/bin/app/build/flutter_assets /usr/share/nginx/html

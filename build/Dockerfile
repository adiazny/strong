# Build the Go Binary.
FROM golang:1.21.4 as build_app
ENV CGO_ENABLED 0
ARG BUILD_REF

# Create the service directory and the copy the module files first and then
# download the dependencies. If this doesn't change, we won't need to do this
# again in future builds.
# RUN mkdir /service
# COPY go.* /service/
# WORKDIR /service
# RUN go mod download

# Copy the source code into the container.
COPY . /app

# Build the binary.
WORKDIR /app/cmd/api
RUN go build

# Run the Go Binary in Alpine.
FROM alpine:3.18
ARG BUILD_DATE
ARG BUILD_REF
RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -h /app -G app -S app
COPY --from=build_app --chown=app:app /app/cmd/api/api /app/strong
WORKDIR /app
USER app
CMD ["./strong"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="strong-app" \
      org.opencontainers.image.authors="Alan Diaz <adiazny@gmail.com>" \
      org.opencontainers.image.revision="${BUILD_REF}" 
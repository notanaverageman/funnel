FROM golang:1.19 AS build

COPY . /src
WORKDIR /src
RUN make release

FROM busybox:1.36.0

COPY --from=build /src/funnel_minimal_linux-amd64 /opt/funnel
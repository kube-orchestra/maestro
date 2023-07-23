FROM registry.access.redhat.com/ubi8/go-toolset

COPY . .
WORKDIR cmd/server
RUN go build -o /tmp/maestro-api -buildvcs=false

CMD ["/tmp/maestro-api"]

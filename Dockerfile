FROM segment/sources-node:3.0.16

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.0.10

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

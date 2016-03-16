FROM segment/sources-node:3.1.0

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

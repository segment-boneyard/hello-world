FROM segment/sources-node:3.0.11

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.0.24

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

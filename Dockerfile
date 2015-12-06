FROM segment/sources-node:3.0.1

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

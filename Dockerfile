FROM segment/sources-node:3.0.2

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

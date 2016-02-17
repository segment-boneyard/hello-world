FROM segment/sources-node:3.0.14

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

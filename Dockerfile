FROM segment/sources-node:3.0.15

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

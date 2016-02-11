FROM segment/sources-node:3.0.12

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

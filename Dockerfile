FROM segment/sources-node:3.5.0

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

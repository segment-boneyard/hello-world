FROM segment/sources-node:3.3.10

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

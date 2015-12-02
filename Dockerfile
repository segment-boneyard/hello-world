FROM segment/sources-node:2.2.1

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

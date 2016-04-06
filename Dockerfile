FROM segment/sources-node:3.2.1

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

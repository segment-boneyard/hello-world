FROM segment/sources-node:3.1.1

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

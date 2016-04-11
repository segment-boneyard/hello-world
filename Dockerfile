FROM segment/sources-node:3.3.2

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

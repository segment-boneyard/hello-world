FROM segment/sources-node:3.3.5

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.3.12

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

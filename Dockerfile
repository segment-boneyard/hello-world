FROM segment/sources-node:3.0.17

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

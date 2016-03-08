FROM segment/sources-node:3.0.25

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

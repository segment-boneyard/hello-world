FROM segment/sources-node:3.0.27

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

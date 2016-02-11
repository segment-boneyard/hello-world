FROM segment/sources-node:3.0.13

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.0.26

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.0.18

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

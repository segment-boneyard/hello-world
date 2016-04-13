FROM segment/sources-node:3.3.9

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

FROM segment/sources-node:3.5.2

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

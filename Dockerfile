FROM segment/sources-node:3.4.6

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

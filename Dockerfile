FROM segment/sources-node:3.3.18

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

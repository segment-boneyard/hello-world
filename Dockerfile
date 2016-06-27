FROM segment/sources-node:v3.5.4

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

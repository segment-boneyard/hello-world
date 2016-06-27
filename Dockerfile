FROM segment/sources-node:v3.5.2

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

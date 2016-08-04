FROM segment/sources-node-v6.3:v4.2.0

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

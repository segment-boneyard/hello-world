FROM segment/sources-node-v6.3:v4.0.8

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

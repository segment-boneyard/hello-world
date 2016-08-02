FROM segment/sources-node-v6.3:v4.1.3

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

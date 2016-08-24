FROM segment/sources-node-v6.3:v4.4.3

COPY . /src

ENTRYPOINT ["/sources", "run", "src/bin/stripe"]

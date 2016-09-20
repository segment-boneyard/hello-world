FROM segment/sources-node-v6.3:v4.6.4

COPY . /src

ENTRYPOINT [ "shifu", "/sources", "run", "src/bin/stripe" ]

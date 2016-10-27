FROM segment/sources-node-v6.3:v4.7.7

COPY . /src

ENTRYPOINT [ "shifu", "/sources", "run", "src/bin/stripe" ]

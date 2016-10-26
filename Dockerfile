FROM segment/sources-node-v6.3:v4.7.6

COPY . /src

ENTRYPOINT [ "shifu", "/sources", "run", "src/bin/stripe" ]

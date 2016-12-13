FROM segment/sources-node-v6.3:v4.8.5

COPY . /src

ENTRYPOINT [ "shifu", "/sources", "run", "src/bin/stripe" ]

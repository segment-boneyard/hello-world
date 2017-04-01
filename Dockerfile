FROM segment/sources-node-v6.3:v4.12.1

COPY . /src

ENTRYPOINT [ "/sources", "run", "src/bin/stripe" ]

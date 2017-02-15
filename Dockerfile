FROM segment/sources-node-v6.3:v4.9.2

COPY . /src

ENTRYPOINT [ "/sources", "run", "src/bin/stripe" ]

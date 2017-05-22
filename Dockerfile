FROM segment/sources-node-v6.3:v4.13.1

COPY . /src

ENTRYPOINT [ "/sources", "run", "src/bin/stripe" ]

FROM segment/sources-node-v6.3:v4.13.2

COPY . /src

ENTRYPOINT [ "/sources", "run", "src/bin/stripe" ]

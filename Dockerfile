FROM segment/sources-node:2.0.1
COPY . /src
ENTRYPOINT ["/sources", "run", "node", "--harmony", "src/bin/stripe.js"]

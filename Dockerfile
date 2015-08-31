FROM segment/sources-node
COPY . /src
ENTRYPOINT ["/sources", "run", "node", "--harmony", "src/bin/stripe.js"]
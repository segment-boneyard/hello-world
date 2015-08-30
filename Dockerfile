FROM segment/sources-node
COPY . /src
CMD ["/sources", "run", "node", "--harmony", "src/bin/stripe.js"]

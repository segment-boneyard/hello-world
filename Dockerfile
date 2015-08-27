FROM segment/sources-node
COPY . /src
CMD ["node", "--harmony", "src/bin/stripe.js"]

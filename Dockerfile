FROM segment/node:v0.12
COPY . /src
ENTRYPOINT ["node"]

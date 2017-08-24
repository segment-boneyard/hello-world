FROM segment/sources:v4.17.1
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

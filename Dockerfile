FROM segment/sources:v4.17.2
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

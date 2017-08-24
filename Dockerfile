FROM segment/sources:v4.17.3
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

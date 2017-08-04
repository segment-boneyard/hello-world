FROM segment/sources:v4.16.17
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

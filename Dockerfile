FROM segment/sources:v4.16.19
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

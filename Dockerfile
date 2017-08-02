FROM segment/sources:v4.16.15
ENV LSP=noop
COPY bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

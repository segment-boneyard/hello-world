FROM segment/sources-build:v1.0.0 as builder
WORKDIR /go/src/github.com/segment-sources/stripe/
COPY . .
RUN govendor install -ldflags '-s -w' .

FROM ${ECR}/sources:v4.18.1
COPY --from=builder /go/bin/stripe /stripe
ENTRYPOINT [ "/sources", "run", "/stripe" ]

FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="Execution Hook Controller"

COPY ./bin/hook hook
ENTRYPOINT ["/hook"]

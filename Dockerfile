FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-pingfederate"]
COPY baton-pingfederate /
FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-dropbox"]
COPY baton-dropbox /
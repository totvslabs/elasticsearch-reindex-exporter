FROM scratch
LABEL maintainer="devops@totvslabs.com"
COPY elasticsearch-reindex-exporter /bin/elasticsearch-reindex-exporter
ENTRYPOINT ["/bin/elasticsearch-reindex-exporter"]
CMD [ "-h" ]

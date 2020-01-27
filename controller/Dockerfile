FROM debian:stable

COPY ./charts/base /charts/base
COPY ./bin/controller-linux /usr/local/bin/helm-demo-controller

ENTRYPOINT [ "/usr/local/bin/helm-demo-controller" ]

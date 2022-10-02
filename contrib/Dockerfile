FROM golang:1.19-alpine3.16@sha256:d475cef843a02575ebdcb1416d98cd76bab90a5ae8bc2cd15f357fc08b6a329f

RUN apk --no-cache add build-base git bash curl openssl npm

RUN curl -fsSL -o ./get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 ./get_helm.sh \
    && ./get_helm.sh \
    && rm ./get_helm.sh

WORKDIR /projects

VOLUME ["/projects"]
EXPOSE 3000

CMD ["bash"]

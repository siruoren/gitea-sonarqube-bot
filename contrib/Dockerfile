FROM golang:1.18-alpine3.16@sha256:7cc62574fcf9c5fb87ad42a9789d5539a6a085971d58ee75dd2ee146cb8a8695

RUN apk --no-cache add build-base git bash curl openssl npm

RUN curl -fsSL -o ./get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 ./get_helm.sh \
    && ./get_helm.sh \
    && rm ./get_helm.sh

WORKDIR /projects

VOLUME ["/projects"]
EXPOSE 3000

CMD ["bash"]

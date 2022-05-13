FROM golang:1.17-alpine3.14

RUN apk --no-cache add build-base git bash curl openssl npm

RUN curl -fsSL -o ./get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 ./get_helm.sh \
    && ./get_helm.sh \
    && rm ./get_helm.sh

WORKDIR /projects

VOLUME ["/projects"]
EXPOSE 3000

CMD ["bash"]

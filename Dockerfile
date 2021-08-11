# Build stage
FROM 230377472753.dkr.ecr.us-east-1.amazonaws.com/bread-golang:master-latest AS builder

# Args
ARG GITHUB_TOKEN

# Variables
ENV APP_PATH=/opt/bread
ENV APP_NAME=shopify_plugin_backend

# Verify GITHUB_TOKEN is set
RUN test -n "$GITHUB_TOKEN" # GITHUB_TOKEN is *required*

# Add the Github token to allow pulling from private, HTTPS-enabled repositories.
RUN echo "machine github.com login $GITHUB_TOKEN" > /root/.netrc \
    && chmod 400 /root/.netrc

# Install chamber
RUN go get github.com/segmentio/chamber

# Install goose
#RUN go get github.com/getbread/goose/cmd/goose
RUN go get bitbucket.org/liamstask/goose/cmd/goose

WORKDIR /src

ADD . $APP_NAME

WORKDIR /src/$APP_NAME/service
RUN go build -o main ./cmd/shopify_plugin_backend

# Build final image
FROM alpine:3.11.3

ENV APP_PATH=/opt/bread
ENV APP_NAME=shopify_plugin_backend

WORKDIR $APP_PATH/$APP_NAME

# uncomment next line if running as docker file outside of slice:
#COPY --from=builder /src/$APP_NAME/deploy/chart/local/config.yaml $APP_PATH/$APP_NAME/deploy/chart/local/config.yaml
# Copy app binary
COPY --from=builder /src/$APP_NAME/service/main $APP_PATH/$APP_NAME/$APP_NAME
# Copy dummy env file
COPY --from=builder /src/$APP_NAME/.dummy_env $APP_PATH/$APP_NAME/.dummy_env
# Copy migration files
COPY --from=builder /src/$APP_NAME/service/internal/storage/migrations $APP_PATH/$APP_NAME/service/internal/storage/migrations
# Copy goose
COPY --from=builder /go/bin/goose /usr/local/bin/goose
# Copy chamber
COPY --from=builder /go/bin/chamber /usr/local/bin/chamber
COPY --from=builder /src/$APP_NAME/service/cmd/shopify_plugin_backend/build $APP_PATH/$APP_NAME/service/cmd/shopify_plugin_backend/build
# Copy root file system files
COPY deploy/rootfs/ /

EXPOSE 8000/tcp
EXPOSE 9000/tcp

CMD ["/opt/bread/shopify_plugin_backend/shopify_plugin_backend"]

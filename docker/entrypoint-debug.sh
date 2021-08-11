#!/bin/sh

#Create goose config file and run db migrations
cd $OUTPUT_DIR
envsubst < ${OUTPUT_DIR}db/dbconf.yml.template > ${OUTPUT_DIR}db/dbconf.yml
goose up

# Run go app
#go get github.com/go-delve/delve/cmd/dlv
dlv exec ${OUTPUT_DIR}${BUILD_PATH}  --headless=true --listen=:40000 --api-version=2 --log -- -env_path=${OUTPUT_DIR}.env

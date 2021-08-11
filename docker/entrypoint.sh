#!/bin/sh

#Create goose config file and run db migrations
cd $OUTPUT_DIR
envsubst < ${OUTPUT_DIR}db/dbconf.yml.template > ${OUTPUT_DIR}db/dbconf.yml
goose up

# Run go app
${OUTPUT_DIR}${BUILD_PATH} -env_path=${OUTPUT_DIR}.env

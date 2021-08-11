#!/bin/sh
echo "hi, I'm the setup sorcerer! I'm like a knock-off install wizard! use me to configure local development with Milton!"
pwd
cd scripts
# verify installed dependencies
echo "checking to make sure all dependencies are good to go..."
if test $NGROK_AUTH_TOKEN; then echo "NGROK_AUTH_TOKEN found"; else { echo "NGROK_AUTH_TOKEN not set, required for set up"; exit 1;}; fi
if [ -x "$(command -v go)" ]; then echo "found go"; else { echo "need to install go"; exit 1;}; fi
if [ -x "$(command -v docker-compose)" ]; then echo "found docker-compose"; else { echo "need to install docker-compose"; exit 1;}; fi
if [ -x "$(command -v ngrok)" ]; then
    echo "found ngrok"; 
else
    echo "let me install ngrok for you..."
    brew install ngrok;
    ngrok authtoken $NGROK_AUTH_TOKEN
fi
echo "let's check your .env file"
if ! test -e ../.env; then 
    cp dev_ens.template ../.env 
    echo "fill out all the empty environmental variables in '.env' and re-run this script "
    exit 0
else
    echo ".env file found!"
fi
#check ngrok auth token

if test $GITHUB_TOKEN; then echo "GITHUB_TOKEN found"; else { echo "GITHUB_TOKEN not set, required for set up"; exit 1;}; fi

test_for_variables=(
TRANSACTION_SERVICE_API_KEY
TRANSACTION_SERVICE_API_SECRET
TRANSACTION_SERVICE_HOST
SHOPIFY_API_KEY
SHOPIFY_SHARED_SECRET
LAUNCHDARKLY_KEY
LAUNCHDARKLY_CLIENT_SIDE_ID
MILTON_ADMIN_SESSION_SECRET
MILTON_OAUTH_CLIENT_ID
MILTON_OAUTH_CLIENT_SECRET
)
source ../.env
for x in "${test_for_variables[@]}"; do
    if ! test "${!x}"; then { echo "need to set environmental variable $x in your .env file"; exit 1;}; fi
done

echo "set up complete!"
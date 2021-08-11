this repo is WIP, currently working on deploying through Chef (need a zip of backend and front end code together) and being deployed separately in the kuberenetes bread pod.

triggering build 

# run as slice project

requires:

- vpn connection
- SDM active
- slice tools
- a slice environment (see this [confluence page](https://breadfinance.atlassian.net/wiki/spaces/ENGINEERING/pages/1527513563/Slice+-+Creating+a+new+slice+environment) for details)

```

#these instructions assume you already installed bread-gateway + provisioned a slice environment
slice chart sync --local # necessary if changing anything in the `deploy folder`
slice start
slice run build
slice run server

#smoke test to make sure it's working
curl -k https://api.<your-slice-environment>.slice.ue2.breadgateway.net/api/shopify-plugin-backend/gateway/tenant
#expect a JSON response
```

# run as bread platform (locally)
since we are using cookie based authentication, and are using no samesite attributes, we need to include a self-signed cert:
one time cert install: 
```
brew install mkcert
mkcert -install
cd <path>/shopify_plugin_backend
mkdir certs
cd certs
mkcert localhost
```
you may need to add the mkcert certificate to your keychain. 

from there you can start the application as follows:
```
make db_boot
make run_be_config
```

at this point, backend will be running on `localhost:8000`

you should be able to run the frontend by changng to that directory and running `npm run dev`

# how to run as bread classic, locally

`./scripts/chef_build/copy_build_github.sh` needs updating for folder name changes

# how to run locally (for bread platform, aka backend only)

note: this information highly subject to change. this is a transitory period in the plugin repo.

1. populate .env file (easiest to do is retrieve a version from Sam K or Sam J) this contains several secrets, and should be transferred securely. retrieve localdev/values.yaml and put in directory `deploy/chart/localdev/config.yaml`. put `.env` in root of project
2. spin up a postgres container, and create a database as specified below:

```
docker pull postgres
docker run --rm --name pg-docker -e POSTGRES_PASSWORD=mooncakes -d -p 5432:5432 -v $HOME/docker/volumes/postgres:/var/lib/postgresql/data postgres
psql -h 127.0.0.1 -p 5432 -U postgres
```

in psql client, add databased by running `CREATE DATABASE shopify`

can be started and stopped subsequently with `docker start pg-docker` and `docker stop pg-docker`

Note: any database listening on localhost:5432 will do the trick.

3. from project root run `build milton/*.go` and then run executable. requires the following envs set:

```
export CONFIG_FILE_NAME=config
export CONFIG_FILE_PATH=./deploy/config/
export ENVIRONMENT=local
```

4. alternatively, if using vs code, launch debugger (see .vscode/launch.json)

this should let you run locally without relying on slice (for faster development)

# how to make a chef_deployable version locally:

- add `SERVE_FE_ROUTES=TRUE` to `.env`
- add `DOCKERFILE=scripts/chef_build/Dockerfile-fullstack` to `.env`
- run `scripts/chef_build/copy_build_github.sh` (fetches and builds frontend from remote, run only as often as frontend is updated on master)
- run `docker-compose build --no-cache`
- run `docker-compose up` you are now running and serving the "chef" version locally

# how to serve separate FE and BE

- in `.env`

```
FE_HOST=http://localhost:7802 #for local dev
```

- spin up BE and FE separately. should work when accessing everything from front end host

## Specifying commit to deploy

Open `travis.yml` and insert a line after line 25

Add the following after the `git clone` statement : `git checkout <commit_sha>`

Example :

```
install:
  - git clone https://github.com/getbread/shopify_plugin_frontend.git
  - git checkout a2994a89eafe7dc8fee814d0c466050d57c729e8
  - cd shopify_plugin_frontend/
  - npm install
```

# below is legacy docs. likely to be Out of date:

## Role

Milton keeps the paperwork in order between Bread & Shopify for our Shopify merchants.

# Install for local development

## prerequisuites

- NGROK_AUTH_TOKEN env set
- GITHUB_TOKEN env set
- Go 1.5+, Docker, and docker-compose installed

## steps

1. run `make install` from project root
2. this will create and populate `.env` Fill out the blank environment variables and then re-run script (you will need another Dev's help with this part), or see example config below
3. you should be good to go now. run `docker-compose up` and verify that all services spin up.

## Config

Create a .env file in the project directory with these values (change if neccessary)

```
ENV=development

HOST=localhost

DB_HOST=localhost
DB_PORT=5432
DB_USER=[username]
DB_PASSWORD=[password]
DB_DATABASE=shopify

DB_MAX_OPEN_CONNS=0
DB_MAX_IDLE_CONNS=0

MILTON_PORT=7800
MILTON_DEBUG_PORT=20522

MILTON_HOST=https://bread-shopify.ngrok.io
BREAD_HOST=https://api-dev.getbread.com
BREAD_HOST_LOCAL=http://localhost:7777
BREAD_HOST_DEVELOPMENT=https://api-dev-sandbox.getbread.com
CHECKOUT_HOST=https://checkout-dev.getbread.com
CHECKOUT_HOST_DEVELOPMENT=https://checkout-dev-sandbox.getbread.com
CHECKOUT_HOST_LOCAL=https://bread-gladys.ngrok.io

SHOPIFY_API_KEY=4fe76d236564a0b1bb8912c9889d9a61
SHOPIFY_SHARED_SECRET=2a83754b2a2eb4474b790a7a94e4d97c

AVALARA_KEY=+QHsEvPZ3MWhG3JVS7t/s3WuXi5u1pVihLskk38RfWz++bklL1J4lw9/jApT/fWp6kogj3XGmqT+QO6V5D5xog==
```

## Build

`make build`. See makefile for other useful commands.

## Testing with development stores

find instructions on setting up a shopify test store [here](https://breadfinance.atlassian.net/l/c/tiXDnGCq)
Shopify Developer Account & Development Store

**_Developer Account URL_**: https://app.shopify.com/services/partners/auth/login

**_Email_**: ask others to have an account created for you

**_Password_**:

**_Payment Gateway Registration Link for Dev_**: available from the developer account

**_App Installation Url_**: http://bread-shopify.ngrok.io/install?shop=<insert_shop_name>

- Run server with ngrok url for webhooks
- Make order on dev [shop](https://bakers-prep.myshopify.com)
- Log into dev shop [admin](https://bakers-prep.myshopify.com/admin)
  - Ensure order was successfully replicated on Shopify store backend
- Call cancel/capture/refund on order on Shopify [admin](https://bakers-prep.myshopify.com/admin) portal
  - Ensure actions successfully update transaction on Bread [sandbox Hawking](https://merchants-sandbox.getbread/com)
  - Note: Bread merchant account tied to Shopify store is Spartan Shop
- Test app settings in the Miltons app [admin console](https://bakers-prep.myshopify.com/admin/apps/4fe76d236564a0b1bb8912c9889d9a61)

## Debugging

to debug, build the docker container in debug mode with `make build-debug`, and then launch the container with `make run`

then you must attach to the debugging instance, which is available on `localhost:40000`

you can connect a CLI dlv debugger with `make debug`
if you are using VS Code, you can add the following to `.vscode/launch.json`, and then launch as normal

```javascript
{
    "version": "0.2.0",
    "configurations": [


        {
                "name": "milton debug",
                "type": "go",
                "request": "attach",
                "mode": "remote",
                "remotePath": "/go/src/github.com/getbread/milton",
                "port": 40000,
                "host": "localhost",
                "showLog": true
        }
    ]
}
```

## Deployment

make a pr in the [chef-repo](https://github.com/getbread/chef/blob/master/data_bags/apps/milton.json) updating with the desired Travis build

## Other Tools used

[ngrok](https://github.com/getbread/base/wiki/Setting-Up-ngrok)

## running DB with ngrok (no docker container)

You need [ngrok](https://ngrok.com/) installed.
Get the team account key from someone if you don't have it.

Ensure db/dbconf.yml looks similar to this

```
development:
  driver: postgres
  open: user=jorgeolivero dbname=shopify sslmode=disable
  table: goose_shopify
```

Then run these command from the correct directories

```bash
$ goose up
```

```bash
$ ./milton.sh
```

Open up another shell for ngrok

```bash
$ ngrok http --subdomain=bread-shopify 7800
```

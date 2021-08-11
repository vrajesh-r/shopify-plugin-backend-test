DIR=/tmp/copy_shopify_frontend
if test -d $DIR
then
    (
        cd $DIR/shopify_plugin_frontend
        git pull
        #uncomment this line to use a remote branch instead of master
        git checkout test-chef-deploy
        npm install
        npm run build; 
    )
else
    mkdir -p $DIR
    (
        cd $DIR
        git clone git@github.com:getbread/shopify_plugin_frontend.git
        cd shopify_plugin_frontend;
        git checkout test-chef-deploy
        #git checkout PG-213-hotfix;
        npm install
        npm run build; 
    )
fi
mkdir -p ./shopify_plugin_backend/build/gateway
rm -r ./shopify_plugin_backend/build/gateway
cp -r $DIR/shopify_plugin_frontend/build/gateway ./shopify_plugin_backend/build/
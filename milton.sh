DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR/milton
grunt
go run main.go -env_path=$DIR/.env

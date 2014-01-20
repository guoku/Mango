go build -o management -x main.go
go build -o jobs/get_taobao_guoku_cats_match -x jobs/get_taobao_guoku_cats_match.go
mkdir upload
cp management upload
cp -r conf upload
cp -r views upload
cp -r static upload
cp -r jobs upload
cp upload/conf/prod.conf upload/conf/app.conf
rsync -avz upload/ jasonzhou@114.113.154.49:/data/www/orange
rm -r upload

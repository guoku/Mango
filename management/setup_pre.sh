go build -o managment -x main.go
mkdir upload
cp managment upload
cp -r conf upload
cp -r views upload
cp -r static upload
cp upload/conf/staging.conf upload/conf/app.conf
rsync -avz upload/ guoku@pre.guoku.com:/data/www/management
rm -r upload

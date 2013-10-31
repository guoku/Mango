go build -o scheduler -x scheduler.go
go build -o management -x main.go
go build -o api_crawler -x apicrawler/api_crawler.go
go build -o jobs/scorer -x jobs/scorer.go
go build -o jobs/get_cats -x jobs/get_taobao_cats.go

mkdir upload
cp management upload
cp api_crawler upload
cp scheduler upload
cp -r conf upload
cp -r views upload
cp -r static upload
cp -r jobs upload
cp server.key upload
cp server.crt upload
cp upload/conf/test.conf upload/conf/app.conf
rsync -avz upload/ gkeng@10.0.1.23:~/scheduler
rm -r upload

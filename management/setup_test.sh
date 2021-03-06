go build -o scheduler -x scheduler.go
go build -o management -x main.go
go build -o jobs/scorer -x jobs/scorer.go
go build -o jobs/get_cats -x jobs/get_taobao_cats.go
go build -o jobs/sync_items -x jobs/sync_items.go
go build -o jobs/item_standarize -x jobs/item_standarize.go
go build -o jobs/get_taobao_guoku_cats_match -x jobs/get_taobao_guoku_cats_match.go
go build -o jobs/word_detector -x jobs/word_detector.go
go build -o jobs/word_detector_imp -x jobs/word_detector_imp.go
go build -o jobs/addtional_extractor -x jobs/addtional_extractor.go
go build -o jobs/oneoff/import_old_shop_data -x jobs/oneoff/import_old_shop_data.go

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

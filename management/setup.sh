go build -o management -x main.go
mkdir upload
cp management upload
cp -r conf upload
cp -r views upload
cp -r static upload
rsync -avz upload/ guoku@10.0.2.50:/data/scheduler
rm -r upload

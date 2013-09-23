#!/bin/bash

find . -name '*.pyc' -exec rm -f {} \; # clean pyc 
rsync -avz --delete --exclude='settings.py' mango/  stxiong@10.0.1.23:/data/www/mango/
#rsync -avz --delete mango/  stxiong@10.0.1.23:/data/www/mango/

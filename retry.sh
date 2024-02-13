#!/bin/bash

s3cmd mb s3://test-bucketname

for i in {1..10000}; do
        echo "vvv i=$i , \$1=$1"
        ./s3example "$1"
        # config s3cmd
        s3cmd get "s3://test-bucketname/test-key$1" /tmp/ --force && s3cmd rm "s3://test-bucketname/test-key$1"
        # failed with 404
        if [[ $? -ne 0 ]]; then
                echo "failed"
                exit 0
        fi
        echo "^^^ i=$i , \$1=$1"
done

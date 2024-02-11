#!/bin/bash

s3cmd mb s3://test-bucketname

for i in {1..10000}; do
        ./s3example
        # config s3cmd
        s3cmd get s3://test-bucketname/test-key /tmp/ --force && s3cmd rm s3://test-bucketname/test-key
        # failed with 404
        if [[ $? -ne 0 ]]; then
                echo "failed"
                exit 0
        fi
done

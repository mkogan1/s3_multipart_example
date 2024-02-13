1. modify `main.go` with ak, sk, endpoint, bucket 和key
2. ```
   dd if=/dev/urandom of=/tmp/test1.txt bs=1M count=20
   dd if=/dev/urandom of=/tmp/test2.txt bs=1M count=30
   ```
     or others
3. install s3cmd add config .s3cfg
4. `go build -o ./s3example ./main.go`
5. `bash retry.sh`
   (in case multiple instances should be run in parallel add a objectname suffix arg, ex: `bash retry.sh $$`)

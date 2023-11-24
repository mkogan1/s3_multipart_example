package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {

	svc := s3.New(session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("accessky", "secrekey", ""),
		Endpoint:         aws.String("http://endpoint:80"),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
	})))
	err := RetryUpload(svc, "test-bucketname", "test-key")
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func RetryUpload(svc *s3.S3, bucket, key string) error {
	var (
		byteSize       int = 101 * 1024 * 1024
		completedParts []*s3.CompletedPart
	)
	buffer := generateRandomBytes(byteSize)

	fmt.Println("start CreateMultipartUpload")
	resp, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		fmt.Println("failed to CreateMultipartUpload ")
		return err
	}
	fmt.Println("success to CreateMultipartUpload")

	fmt.Println("start UploadPart PartNumber 1")
	// first update 8 MB
	fileBytes := buffer[0 : 8*1024*1024]
	uploadResult, err := svc.UploadPart(&s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        &bucket,
		Key:           &key,
		PartNumber:    aws.Int64(int64(1)),
		UploadId:      &*resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	})

	if err != nil {
		fmt.Println("failed to UploadPart PartNumber 1")
		return err
	}

	etag := *uploadResult.ETag
	// fmt.Println("etag: %s\n", etag)
	completedParts = append(completedParts, &s3.CompletedPart{
		ETag:       &etag,
		PartNumber: aws.Int64(int64(1)),
	})
	fmt.Println("success to UploadPart PartNumber 1")

	fileBytes2 := buffer[1*1024*1024:]

	// dd if=/dev/urandom of=/tmp/test1.txt bs=1M count=20
	f1, err := os.Open("/tmp/test1.txt")
	if err != nil {
		fmt.Println("failed to CompleteMultipartUpload")
		return err
	}
	defer f1.Close()

	// dd if=/dev/urandom of=/tmp/test2.txt bs=1M count=30
	f2, err := os.Open("/tmp/test2.txt")
	if err != nil {
		fmt.Println("failed to CompleteMultipartUpload")
		return err
	}
	defer f2.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	var f func(bytess io.ReadSeeker) = func(bytess io.ReadSeeker) {

		var buf = make([]byte, 64)
		var stk = buf[:runtime.Stack(buf, false)]
		fmt.Println("start UploadPart PartNumber 2, goroutine id " + string(stk))
		// second part
		uploadResult2, err := svc.UploadPart(&s3.UploadPartInput{
			Body:          bytess,
			Bucket:        &bucket,
			Key:           &key,
			PartNumber:    aws.Int64(int64(2)),
			UploadId:      &*resp.UploadId,
			ContentLength: aws.Int64(int64(100 * 1024 * 1024)),
		})
		if err != nil {
			fmt.Println("failed to UploadPart PartNumber 2, goroutine id " + string(stk) + err.Error())
			return
		}
		fmt.Println("success to UploadPart PartNumber 2, now append, goroutine id " + string(stk))
		completedParts = append(completedParts, &s3.CompletedPart{
			ETag:       &*uploadResult2.ETag,
			PartNumber: aws.Int64(int64(2)),
		})
		wg.Done()
	}

	go f(f1)
	go f(f2)
	go f(bytes.NewReader(fileBytes2))

	wg.Wait()
	fmt.Println("CompleteMultipartUpload start")
	_, err = svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   &bucket,
		Key:      &key,
		UploadId: &*resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		fmt.Println("failed to CompleteMultipartUpload")
		return err
	}
	fmt.Println("CompleteMultipartUpload success ")
	panic("success upload")
}

func generateRandomBytes(size int) []byte {
	rndSrc := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(rndSrc)
	random := make([]byte, size)
	var (
		err   error
		n     = 0
		retry = 3
		i     = 0
	)
	for ; i < retry && n != size; i++ {
		n, err = io.ReadFull(rng, random)
	}
	if err != nil {
		return nil
	}
	return random
}

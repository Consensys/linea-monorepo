package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func getAWSConfig() aws.Config {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-2"))
	if err != nil {
		fmt.Println("unable to load AWS config: see https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials, " + err.Error())
		os.Exit(-1)
	}

	return cfg
}

func downloadFromS3(bucket, key string) (io.ReadCloser, error) {

	// Create an Amazon S3 service s3Client
	s3Client := getAWSS3Client()

	fmt.Printf("downloading s3://%s/%s ...\n", bucket, key)
	out, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func getAWSS3Client() *s3.Client {
	return s3.NewFromConfig(getAWSConfig())
}

func uploadToS3(s3client *s3.Client, bucketName string, objectKey string, fileName string) error {
	fmt.Printf("uploading %s to s3://%s/%s ...\n", fileName, bucketName, objectKey)
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Couldn't open file %v to upload. Here's why: %v\n", fileName, err)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	checkError(err)

	// Above 4Gb, the file is considered as a large file
	if fileinfo.Size() > 4_000_000_000 {
		fmt.Printf("file is %d bytes large, using multipart upload", fileinfo.Size())
		buf := make([]byte, fileinfo.Size())
		_, err := file.Read(buf)
		checkError(err)
		return uploadLargeObject(s3client, bucketName, objectKey, buf)
	}

	_, err = s3client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
			fileName, bucketName, objectKey, err)
	}
	return nil
}

// UploadLargeObject uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func uploadLargeObject(s3Client *s3.Client, bucketName string, objectKey string, largeObject []byte) error {
	largeBuffer := bytes.NewReader(largeObject)
	var partMiBs int64 = 10
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	fmt.Printf("built the uploader\n")
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:            aws.String(bucketName),
		Key:               aws.String(objectKey),
		Body:              largeBuffer,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
	})
	fmt.Printf("successfull uploaded %v\n", objectKey)
	if err != nil {
		log.Printf("Couldn't upload large object to %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}

	return err
}

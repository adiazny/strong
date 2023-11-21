package store

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"golang.org/x/oauth2"
)

type S3Object struct {
	Log *log.Logger
	*s3.Client
	BucketName string
	ObjectKey  string
}

func (s *S3Object) GetToken() (*oauth2.Token, error) {
	result, err := s.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.ObjectKey),
	})
	if err != nil {
		s.Log.Printf("error getting object %v:%v. %v\n", s.BucketName, s.ObjectKey, err)
		return nil, err
	}

	defer result.Body.Close()

	var t *oauth2.Token

	data := json.NewDecoder(result.Body)

	return t, data.Decode(&t)
}

func (s *S3Object) SetToken(token *oauth2.Token) error {
	data, err := json.Marshal(&token)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(data)

	_, err = s.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.ObjectKey),
		Body:   buf,
	})
	if err != nil {
		s.Log.Printf("error uploading token to %v:%v. %v\n",
			s.BucketName, s.ObjectKey, err)
		return err
	}

	return nil
}

func (s *S3Object) TokenNotPresent() bool {
	_, err := s.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(s.ObjectKey),
	})

	return err != nil
}

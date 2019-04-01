package backend

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo"
)

const (
	nifcloudStorageEndpoint = "https://jp-east-2.storage.api.nifcloud.com"
	nifcloudStorageRegion   = "jp-east-2"
)

var (
	bucketName      = os.Getenv("NIFCLOUD_STORAGE_BUCKET_NAME")
	accessKeyID     = os.Getenv("NIFCLOUD_ACCESS_KEY_ID")
	secretAccessKey = os.Getenv("NIFCLOUD_SECRET_ACCESS_KEY")

	nifcloudService *s3.S3
)

func init() {
	defaultResolver := endpoints.NewDefaultResolver()
	nifcloudResolver := func(service, region string) (aws.Endpoint, error) {
		if service == endpoints.S3ServiceID {
			return aws.Endpoint{
				SigningRegion: nifcloudStorageRegion,
				URL:           nifcloudStorageEndpoint,
			}, nil
		}
		return defaultResolver.ResolveEndpoint(service, region)
	}

	cfg := defaults.Config()
	cfg.EndpointResolver = aws.EndpointResolverFunc(nifcloudResolver)
	cfg.Credentials = aws.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
	nifcloudService = s3.New(cfg)
}

func CreateMultipartUpload(c echo.Context) error {
	fileName := c.QueryParam("fileName")

	req := nifcloudService.CreateMultipartUploadRequest(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	res, err := req.Send()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, *res.UploadId)
}

func GetUploadURL(c echo.Context) error {
	fileName := c.QueryParam("fileName")
	uploadID := c.QueryParam("uploadId")
	partNumber, _ := strconv.ParseInt(c.QueryParam("partNumber"), 10, 64)

	req := nifcloudService.UploadPartRequest(&s3.UploadPartInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(fileName),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int64(partNumber),
	})

	str, err := req.Presign(15 * time.Minute)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, str)
}

func CompleteMultipartUpload(c echo.Context) error {
	type Body struct {
		FileName string             `json:fileName`
		UploadID string             `json:uploadId`
		Parts    []s3.CompletedPart `json:parts`
	}

	body := &Body{}
	if err := c.Bind(body); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	req := nifcloudService.CompleteMultipartUploadRequest(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(body.FileName),
		UploadId: aws.String(body.UploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: body.Parts,
		},
	})
	_, err := req.Send()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "Complete")
}

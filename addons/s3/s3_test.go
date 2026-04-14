package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func TestNewRequiresRegion(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), Config{})
	if err == nil || !strings.Contains(err.Error(), "region is required") {
		t.Fatalf("expected missing region error, got %v", err)
	}
}

func TestNewUsesExplicitConfig(t *testing.T) {
	t.Parallel()

	originalLoad := loadAWSConfig
	originalNewClient := newS3Client
	originalNewPresign := newPresignClient
	t.Cleanup(func() {
		loadAWSConfig = originalLoad
		newS3Client = originalNewClient
		newPresignClient = originalNewPresign
	})

	var capturedLoadOptions awsconfig.LoadOptions
	loadAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		for _, optFn := range optFns {
			if err := optFn(&capturedLoadOptions); err != nil {
				return aws.Config{}, err
			}
		}

		return aws.Config{
			Region:      capturedLoadOptions.Region,
			Credentials: capturedLoadOptions.Credentials,
		}, nil
	}

	var capturedS3Options awss3.Options
	newS3Client = func(cfg aws.Config, optFns ...func(*awss3.Options)) *awss3.Client {
		for _, optFn := range optFns {
			optFn(&capturedS3Options)
		}
		return awss3.NewFromConfig(cfg)
	}

	var presignCalled bool
	newPresignClient = func(client *awss3.Client, optFns ...func(*awss3.PresignOptions)) *awss3.PresignClient {
		presignCalled = true
		return awss3.NewPresignClient(client, optFns...)
	}

	service, err := New(context.Background(), Config{
		Region:         "ap-southeast-1",
		Endpoint:       "play.min.io:9000",
		AccessKey:      "ak",
		SecretKey:      "sk",
		SessionToken:   "session",
		Bucket:         "assets",
		UseSSL:         false,
		ForcePathStyle: true,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got, want := capturedLoadOptions.Region, "ap-southeast-1"; got != want {
		t.Fatalf("region mismatch: got %q want %q", got, want)
	}

	creds, err := capturedLoadOptions.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve credentials failed: %v", err)
	}
	if got, want := creds.AccessKeyID, "ak"; got != want {
		t.Fatalf("access key mismatch: got %q want %q", got, want)
	}
	if got, want := creds.SecretAccessKey, "sk"; got != want {
		t.Fatalf("secret key mismatch: got %q want %q", got, want)
	}
	if got, want := creds.SessionToken, "session"; got != want {
		t.Fatalf("session token mismatch: got %q want %q", got, want)
	}

	if got, want := aws.ToString(capturedS3Options.BaseEndpoint), "http://play.min.io:9000"; got != want {
		t.Fatalf("base endpoint mismatch: got %q want %q", got, want)
	}
	if !capturedS3Options.UsePathStyle {
		t.Fatalf("expected path style enabled")
	}
	if !presignCalled {
		t.Fatalf("expected presign client creation")
	}
	if service.Client() == nil {
		t.Fatalf("expected raw client")
	}
	if service.PresignClient() == nil {
		t.Fatalf("expected presign client")
	}
	if got, want := service.cfg.PresignExpire, defaultPresignExpire; got != want {
		t.Fatalf("default presign expire mismatch: got %s want %s", got, want)
	}
}

func TestNewUsesAnonymousCredentialsWhenExplicitKeysAreAbsent(t *testing.T) {
	t.Parallel()

	originalLoad := loadAWSConfig
	t.Cleanup(func() {
		loadAWSConfig = originalLoad
	})

	var capturedLoadOptions awsconfig.LoadOptions
	loadAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		for _, optFn := range optFns {
			if err := optFn(&capturedLoadOptions); err != nil {
				return aws.Config{}, err
			}
		}
		return aws.Config{
			Region:      capturedLoadOptions.Region,
			Credentials: capturedLoadOptions.Credentials,
		}, nil
	}

	if _, err := New(context.Background(), Config{
		Region:   "ap-southeast-1",
		Endpoint: "example.com",
		Bucket:   "assets",
	}); err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, ok := capturedLoadOptions.Credentials.(aws.AnonymousCredentials); !ok {
		t.Fatalf("expected anonymous credentials provider, got %T", capturedLoadOptions.Credentials)
	}
}

func TestNewRejectsPartialStaticCredentials(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), Config{
		Region:    "ap-southeast-1",
		AccessKey: "ak",
	})
	if err == nil || !strings.Contains(err.Error(), "access key and secret key") {
		t.Fatalf("expected partial credential error, got %v", err)
	}
}

func TestNewRejectsSessionTokenWithoutKeyPair(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), Config{
		Region:       "ap-southeast-1",
		SessionToken: "session",
	})
	if err == nil || !strings.Contains(err.Error(), "session token requires access key and secret key") {
		t.Fatalf("expected session token validation error, got %v", err)
	}
}

func TestServiceRequiresBucketButBucketWrapperCanOverride(t *testing.T) {
	t.Parallel()

	fakeClient := &fakeObjectAPI{
		headObjectFunc: func(ctx context.Context, input *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			if got, want := aws.ToString(input.Bucket), "images"; got != want {
				t.Fatalf("bucket mismatch: got %q want %q", got, want)
			}
			return &awss3.HeadObjectOutput{ETag: aws.String(`"etag"`)}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "")

	if _, err := service.HeadObject(context.Background(), "logo.png"); err == nil || !strings.Contains(err.Error(), "bucket is required") {
		t.Fatalf("expected bucket required error, got %v", err)
	}

	info, err := service.Bucket("images").HeadObject(context.Background(), "logo.png")
	if err != nil {
		t.Fatalf("Bucket.HeadObject returned error: %v", err)
	}
	if got, want := info.Bucket, "images"; got != want {
		t.Fatalf("info bucket mismatch: got %q want %q", got, want)
	}
}

func TestWithBucketOverridesDefaultBucket(t *testing.T) {
	t.Parallel()

	fakeClient := &fakeObjectAPI{
		headObjectFunc: func(ctx context.Context, input *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			if got, want := aws.ToString(input.Bucket), "override"; got != want {
				t.Fatalf("bucket mismatch: got %q want %q", got, want)
			}
			return &awss3.HeadObjectOutput{}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "default")
	if _, err := service.HeadObject(context.Background(), "file.txt", WithBucket("override")); err != nil {
		t.Fatalf("HeadObject returned error: %v", err)
	}
}

func TestUploadBytesDownloadBytesAndHeadObject(t *testing.T) {
	t.Parallel()

	now := time.Unix(1700000000, 0).UTC()
	fakeClient := &fakeObjectAPI{
		putObjectFunc: func(ctx context.Context, input *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			body, err := io.ReadAll(input.Body)
			if err != nil {
				t.Fatalf("read body failed: %v", err)
			}
			if got, want := string(body), "hello"; got != want {
				t.Fatalf("body mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.Bucket), "assets"; got != want {
				t.Fatalf("bucket mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ContentType), "text/plain; charset=utf-8"; got != want {
				t.Fatalf("content type mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.CacheControl), "no-cache"; got != want {
				t.Fatalf("cache control mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ContentDisposition), `inline; filename="hello.txt"`; got != want {
				t.Fatalf("content disposition mismatch: got %q want %q", got, want)
			}
			if got, want := input.Metadata, map[string]string{"source": "test"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("metadata mismatch: got %#v want %#v", got, want)
			}
			return &awss3.PutObjectOutput{
				ETag:      aws.String(`"put-etag"`),
				VersionId: aws.String("v1"),
			}, nil
		},
		getObjectFunc: func(ctx context.Context, input *awss3.GetObjectInput, _ ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
			if got, want := aws.ToString(input.Bucket), "assets"; got != want {
				t.Fatalf("download bucket mismatch: got %q want %q", got, want)
			}
			return &awss3.GetObjectOutput{
				Body:          io.NopCloser(strings.NewReader("hello")),
				ETag:          aws.String(`"get-etag"`),
				ContentType:   aws.String("text/plain"),
				ContentLength: aws.Int64(5),
				LastModified:  aws.Time(now),
				Metadata:      map[string]string{"kind": "text"},
			}, nil
		},
		headObjectFunc: func(ctx context.Context, input *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			return &awss3.HeadObjectOutput{
				ETag:          aws.String(`"head-etag"`),
				ContentType:   aws.String("text/plain"),
				ContentLength: aws.Int64(5),
				LastModified:  aws.Time(now),
				Metadata:      map[string]string{"kind": "text"},
			}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "assets")

	uploadResult, err := service.UploadBytes(
		context.Background(),
		"hello.txt",
		[]byte("hello"),
		WithContentType("text/plain; charset=utf-8"),
		WithCacheControl("no-cache"),
		WithContentDisposition(`inline; filename="hello.txt"`),
		WithMetadata(map[string]string{"source": "test"}),
	)
	if err != nil {
		t.Fatalf("UploadBytes returned error: %v", err)
	}
	if got, want := uploadResult.ETag, `"put-etag"`; got != want {
		t.Fatalf("upload etag mismatch: got %q want %q", got, want)
	}
	if got, want := uploadResult.VersionID, "v1"; got != want {
		t.Fatalf("upload version mismatch: got %q want %q", got, want)
	}

	data, err := service.DownloadBytes(context.Background(), "hello.txt")
	if err != nil {
		t.Fatalf("DownloadBytes returned error: %v", err)
	}
	if got, want := string(data), "hello"; got != want {
		t.Fatalf("download data mismatch: got %q want %q", got, want)
	}

	info, err := service.HeadObject(context.Background(), "hello.txt")
	if err != nil {
		t.Fatalf("HeadObject returned error: %v", err)
	}
	if got, want := info.ETag, `"head-etag"`; got != want {
		t.Fatalf("head etag mismatch: got %q want %q", got, want)
	}
	if got, want := info.ContentLength, int64(5); got != want {
		t.Fatalf("head content length mismatch: got %d want %d", got, want)
	}
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	tempFile, err := os.CreateTemp(t.TempDir(), "upload-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString("from-file"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	fakeClient := &fakeObjectAPI{
		putObjectFunc: func(ctx context.Context, input *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			body, err := io.ReadAll(input.Body)
			if err != nil {
				t.Fatalf("read body failed: %v", err)
			}
			if got, want := string(body), "from-file"; got != want {
				t.Fatalf("file body mismatch: got %q want %q", got, want)
			}
			return &awss3.PutObjectOutput{}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "files")
	if _, err := service.UploadFile(context.Background(), "notes.txt", tempFile.Name()); err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
}

func TestDownloadStreamReturnsMetadata(t *testing.T) {
	t.Parallel()

	now := time.Unix(1700000001, 0).UTC()
	fakeClient := &fakeObjectAPI{
		getObjectFunc: func(ctx context.Context, input *awss3.GetObjectInput, _ ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
			return &awss3.GetObjectOutput{
				Body:          io.NopCloser(strings.NewReader("stream")),
				ETag:          aws.String(`"stream-etag"`),
				ContentType:   aws.String("application/octet-stream"),
				ContentLength: aws.Int64(6),
				LastModified:  aws.Time(now),
				Metadata:      map[string]string{"a": "b"},
			}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "stream-bucket")
	body, info, err := service.DownloadStream(context.Background(), "payload.bin")
	if err != nil {
		t.Fatalf("DownloadStream returned error: %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if got, want := string(data), "stream"; got != want {
		t.Fatalf("stream data mismatch: got %q want %q", got, want)
	}
	if got, want := info.Bucket, "stream-bucket"; got != want {
		t.Fatalf("info bucket mismatch: got %q want %q", got, want)
	}
	if got, want := info.Key, "payload.bin"; got != want {
		t.Fatalf("info key mismatch: got %q want %q", got, want)
	}
	if got, want := info.Metadata, map[string]string{"a": "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("metadata mismatch: got %#v want %#v", got, want)
	}
}

func TestDeleteObjectTreatsNotFoundAsSuccess(t *testing.T) {
	t.Parallel()

	fakeClient := &fakeObjectAPI{
		deleteObjectFunc: func(ctx context.Context, input *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
			return nil, fakeAPIError{code: "NoSuchKey", message: "missing"}
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "assets")
	if err := service.DeleteObject(context.Background(), "missing.txt"); err != nil {
		t.Fatalf("DeleteObject returned error: %v", err)
	}
}

func TestHeadObjectMapsNotFound(t *testing.T) {
	t.Parallel()

	fakeClient := &fakeObjectAPI{
		headObjectFunc: func(ctx context.Context, input *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			return nil, fakeAPIError{code: "NotFound", message: "missing"}
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "assets")
	_, err := service.HeadObject(context.Background(), "missing.txt")
	if !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("expected ErrObjectNotFound, got %v", err)
	}
}

func TestPresignOperations(t *testing.T) {
	t.Parallel()

	fakePresign := &fakePresignAPI{
		presignGetObjectFunc: func(ctx context.Context, input *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
			presignOpts := awss3.PresignOptions{}
			for _, optFn := range optFns {
				optFn(&presignOpts)
			}
			if got, want := aws.ToString(input.Bucket), "assets"; got != want {
				t.Fatalf("presign get bucket mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ResponseContentType), "text/html"; got != want {
				t.Fatalf("response content type mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ResponseCacheControl), "max-age=60"; got != want {
				t.Fatalf("response cache control mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ResponseContentDisposition), "inline"; got != want {
				t.Fatalf("response content disposition mismatch: got %q want %q", got, want)
			}
			if got, want := presignOpts.Expires, 2*time.Minute; got != want {
				t.Fatalf("presign get expire mismatch: got %s want %s", got, want)
			}
			return &awsv4.PresignedHTTPRequest{
				URL:          "https://example.com/get",
				SignedHeader: http.Header{"X-Test": []string{"1"}},
			}, nil
		},
		presignPutObjectFunc: func(ctx context.Context, input *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
			presignOpts := awss3.PresignOptions{}
			for _, optFn := range optFns {
				optFn(&presignOpts)
			}
			if got, want := presignOpts.Expires, 10*time.Minute; got != want {
				t.Fatalf("presign put expire mismatch: got %s want %s", got, want)
			}
			if got, want := aws.ToString(input.ContentType), "application/json"; got != want {
				t.Fatalf("put content type mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.CacheControl), "no-store"; got != want {
				t.Fatalf("put cache control mismatch: got %q want %q", got, want)
			}
			if got, want := aws.ToString(input.ContentDisposition), "attachment"; got != want {
				t.Fatalf("put content disposition mismatch: got %q want %q", got, want)
			}
			if got, want := input.Metadata, map[string]string{"trace": "1"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("put metadata mismatch: got %#v want %#v", got, want)
			}
			return &awsv4.PresignedHTTPRequest{
				URL:          "https://example.com/put",
				SignedHeader: http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		},
		presignDeleteObjectFunc: func(ctx context.Context, input *awss3.DeleteObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
			presignOpts := awss3.PresignOptions{}
			for _, optFn := range optFns {
				optFn(&presignOpts)
			}
			if got, want := presignOpts.Expires, 10*time.Minute; got != want {
				t.Fatalf("presign delete expire mismatch: got %s want %s", got, want)
			}
			if got, want := aws.ToString(input.Bucket), "assets"; got != want {
				t.Fatalf("presign delete bucket mismatch: got %q want %q", got, want)
			}
			return &awsv4.PresignedHTTPRequest{
				URL:          "https://example.com/delete",
				SignedHeader: http.Header{"X-Delete": []string{"1"}},
			}, nil
		},
	}

	service := newTestService(&fakeObjectAPI{}, fakePresign, "assets")
	service.cfg.PresignExpire = 10 * time.Minute

	getURL, getHeaders, err := service.PresignGet(
		context.Background(),
		"index.html",
		WithContentType("text/html"),
		WithCacheControl("max-age=60"),
		WithContentDisposition("inline"),
		WithExpire(2*time.Minute),
	)
	if err != nil {
		t.Fatalf("PresignGet returned error: %v", err)
	}
	if got, want := getURL, "https://example.com/get"; got != want {
		t.Fatalf("presign get url mismatch: got %q want %q", got, want)
	}
	if got, want := getHeaders.Get("X-Test"), "1"; got != want {
		t.Fatalf("presign get header mismatch: got %q want %q", got, want)
	}

	putURL, putHeaders, err := service.PresignPut(
		context.Background(),
		"payload.json",
		WithContentType("application/json"),
		WithCacheControl("no-store"),
		WithContentDisposition("attachment"),
		WithMetadata(map[string]string{"trace": "1"}),
	)
	if err != nil {
		t.Fatalf("PresignPut returned error: %v", err)
	}
	if got, want := putURL, "https://example.com/put"; got != want {
		t.Fatalf("presign put url mismatch: got %q want %q", got, want)
	}
	if got, want := putHeaders.Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("presign put header mismatch: got %q want %q", got, want)
	}

	deleteURL, deleteHeaders, err := service.PresignDelete(context.Background(), "payload.json")
	if err != nil {
		t.Fatalf("PresignDelete returned error: %v", err)
	}
	if got, want := deleteURL, "https://example.com/delete"; got != want {
		t.Fatalf("presign delete url mismatch: got %q want %q", got, want)
	}
	if got, want := deleteHeaders.Get("X-Delete"), "1"; got != want {
		t.Fatalf("presign delete header mismatch: got %q want %q", got, want)
	}
}

func TestBucketWrapperPreservesRawNameAndAllowsOverride(t *testing.T) {
	t.Parallel()

	fakeClient := &fakeObjectAPI{
		headObjectFunc: func(ctx context.Context, input *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			if got, want := aws.ToString(input.Bucket), "override"; got != want {
				t.Fatalf("bucket mismatch: got %q want %q", got, want)
			}
			return &awss3.HeadObjectOutput{}, nil
		},
	}

	service := newTestService(fakeClient, &fakePresignAPI{}, "default")
	bucket := service.Bucket("images")
	if got, want := bucket.RawName(), "images"; got != want {
		t.Fatalf("raw bucket name mismatch: got %q want %q", got, want)
	}

	if _, err := bucket.HeadObject(context.Background(), "logo.png", WithBucket("override")); err != nil {
		t.Fatalf("bucket override returned error: %v", err)
	}
}

func newTestService(client objectAPI, signer presignAPI, bucket string) *Service {
	rawClient := awss3.NewFromConfig(aws.Config{
		Region:      "ap-southeast-1",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("ak", "sk", "")),
	})

	return &Service{
		cfg:     Config{Region: "ap-southeast-1", Bucket: bucket, PresignExpire: defaultPresignExpire},
		raw:     rawClient,
		presign: awss3.NewPresignClient(rawClient),
		client:  client,
		signer:  signer,
		bucket:  bucket,
	}
}

type fakeObjectAPI struct {
	headObjectFunc   func(context.Context, *awss3.HeadObjectInput, ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error)
	deleteObjectFunc func(context.Context, *awss3.DeleteObjectInput, ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
	getObjectFunc    func(context.Context, *awss3.GetObjectInput, ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
	putObjectFunc    func(context.Context, *awss3.PutObjectInput, ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
}

func (f *fakeObjectAPI) HeadObject(ctx context.Context, input *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
	if f.headObjectFunc == nil {
		return nil, errors.New("unexpected HeadObject call")
	}
	return f.headObjectFunc(ctx, input, optFns...)
}

func (f *fakeObjectAPI) DeleteObject(ctx context.Context, input *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
	if f.deleteObjectFunc == nil {
		return nil, errors.New("unexpected DeleteObject call")
	}
	return f.deleteObjectFunc(ctx, input, optFns...)
}

func (f *fakeObjectAPI) GetObject(ctx context.Context, input *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
	if f.getObjectFunc == nil {
		return nil, errors.New("unexpected GetObject call")
	}
	return f.getObjectFunc(ctx, input, optFns...)
}

func (f *fakeObjectAPI) PutObject(ctx context.Context, input *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	if f.putObjectFunc == nil {
		return nil, errors.New("unexpected PutObject call")
	}
	return f.putObjectFunc(ctx, input, optFns...)
}

type fakePresignAPI struct {
	presignGetObjectFunc    func(context.Context, *awss3.GetObjectInput, ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
	presignPutObjectFunc    func(context.Context, *awss3.PutObjectInput, ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
	presignDeleteObjectFunc func(context.Context, *awss3.DeleteObjectInput, ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
}

func (f *fakePresignAPI) PresignGetObject(ctx context.Context, input *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
	if f.presignGetObjectFunc == nil {
		return nil, errors.New("unexpected PresignGetObject call")
	}
	return f.presignGetObjectFunc(ctx, input, optFns...)
}

func (f *fakePresignAPI) PresignPutObject(ctx context.Context, input *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
	if f.presignPutObjectFunc == nil {
		return nil, errors.New("unexpected PresignPutObject call")
	}
	return f.presignPutObjectFunc(ctx, input, optFns...)
}

func (f *fakePresignAPI) PresignDeleteObject(ctx context.Context, input *awss3.DeleteObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error) {
	if f.presignDeleteObjectFunc == nil {
		return nil, errors.New("unexpected PresignDeleteObject call")
	}
	return f.presignDeleteObjectFunc(ctx, input, optFns...)
}

type fakeAPIError struct {
	code    string
	message string
}

func (e fakeAPIError) Error() string {
	return e.code + ": " + e.message
}

func (e fakeAPIError) ErrorCode() string {
	return e.code
}

func (e fakeAPIError) ErrorMessage() string {
	return e.message
}

func (e fakeAPIError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

var _ smithy.APIError = fakeAPIError{}
var _ ObjectOption = bucketOption{}
var _ PutOption = bucketOption{}
var _ PutOption = contentTypeOption{}
var _ PutOption = cacheControlOption{}
var _ PutOption = contentDispositionOption{}
var _ PutOption = metadataOption{}
var _ PresignOption = bucketOption{}
var _ PresignOption = contentTypeOption{}
var _ PresignOption = cacheControlOption{}
var _ PresignOption = contentDispositionOption{}
var _ PresignOption = metadataOption{}
var _ PresignOption = expireOption{}
var _ objectAPI = (*fakeObjectAPI)(nil)
var _ presignAPI = (*fakePresignAPI)(nil)

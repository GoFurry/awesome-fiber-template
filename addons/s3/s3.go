package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var ErrObjectNotFound = errors.New("s3 object not found")

const defaultPresignExpire = 15 * time.Minute

type Config struct {
	Region         string
	Endpoint       string
	AccessKey      string
	SecretKey      string
	SessionToken   string
	Bucket         string
	UseSSL         bool
	ForcePathStyle bool
	PresignExpire  time.Duration
}

type Service struct {
	cfg     Config
	raw     *awss3.Client
	presign *awss3.PresignClient
	client  objectAPI
	signer  presignAPI
	bucket  string
}

type Bucket struct {
	service *Service
	name    string
}

type ObjectInfo struct {
	Bucket        string
	Key           string
	ETag          string
	ContentType   string
	ContentLength int64
	LastModified  time.Time
	Metadata      map[string]string
}

type UploadResult struct {
	Bucket    string
	Key       string
	ETag      string
	VersionID string
}

type PutOptions struct {
	Bucket             string
	ContentType        string
	CacheControl       string
	ContentDisposition string
	Metadata           map[string]string
}

type PresignOptions struct {
	Bucket             string
	Expire             time.Duration
	ContentType        string
	CacheControl       string
	ContentDisposition string
	Metadata           map[string]string
}

type ObjectOptions struct {
	Bucket string
}

type ObjectOption interface {
	applyObject(*ObjectOptions)
}

type PutOption interface {
	applyPut(*PutOptions)
}

type PresignOption interface {
	applyPresign(*PresignOptions)
}

type bucketOption struct {
	name string
}

type contentTypeOption struct {
	value string
}

type cacheControlOption struct {
	value string
}

type contentDispositionOption struct {
	value string
}

type metadataOption struct {
	values map[string]string
}

type expireOption struct {
	value time.Duration
}

type objectAPI interface {
	HeadObject(ctx context.Context, params *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
	GetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
}

type presignAPI interface {
	PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
	PresignDeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.PresignOptions)) (*awsv4.PresignedHTTPRequest, error)
}

var loadAWSConfig = awsconfig.LoadDefaultConfig

var newS3Client = func(cfg aws.Config, optFns ...func(*awss3.Options)) *awss3.Client {
	return awss3.NewFromConfig(cfg, optFns...)
}

var newPresignClient = func(client *awss3.Client, optFns ...func(*awss3.PresignOptions)) *awss3.PresignClient {
	return awss3.NewPresignClient(client, optFns...)
}

func New(ctx context.Context, cfg Config) (*Service, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	awsCfg, err := buildAWSConfig(ctxOrBackground(ctx), normalized)
	if err != nil {
		return nil, err
	}

	rawClient := newS3Client(awsCfg, buildClientOptionFns(normalized)...)
	rawPresign := newPresignClient(rawClient)

	return &Service{
		cfg:     normalized,
		raw:     rawClient,
		presign: rawPresign,
		client:  rawClient,
		signer:  rawPresign,
		bucket:  normalized.Bucket,
	}, nil
}

func (s *Service) Client() *awss3.Client {
	if s == nil {
		return nil
	}
	return s.raw
}

func (s *Service) PresignClient() *awss3.PresignClient {
	if s == nil {
		return nil
	}
	return s.presign
}

func (s *Service) Bucket(name ...string) *Bucket {
	if s == nil {
		return &Bucket{}
	}
	if len(name) == 0 {
		return &Bucket{service: s, name: strings.TrimSpace(s.bucket)}
	}
	return &Bucket{service: s, name: strings.TrimSpace(name[0])}
}

func (s *Service) HeadObject(ctx context.Context, key string, opts ...ObjectOption) (*ObjectInfo, error) {
	service, err := ensureService(s)
	if err != nil {
		return nil, err
	}

	objectOpts, bucket, err := service.resolveObjectOptions(opts...)
	if err != nil {
		return nil, err
	}

	input, err := buildHeadInput(bucket, key, objectOpts)
	if err != nil {
		return nil, err
	}

	output, err := service.client.HeadObject(ctxOrBackground(ctx), input)
	if err != nil {
		return nil, classifyS3Error(err)
	}

	return mapHeadObject(bucket, key, output), nil
}

func (s *Service) DeleteObject(ctx context.Context, key string, opts ...ObjectOption) error {
	service, err := ensureService(s)
	if err != nil {
		return err
	}

	objectOpts, bucket, err := service.resolveObjectOptions(opts...)
	if err != nil {
		return err
	}

	input, err := buildDeleteInput(bucket, key, objectOpts)
	if err != nil {
		return err
	}

	_, err = service.client.DeleteObject(ctxOrBackground(ctx), input)
	if err != nil {
		classified := classifyS3Error(err)
		if !errors.Is(classified, ErrObjectNotFound) {
			return classified
		}
	}

	return nil
}

func (s *Service) DownloadBytes(ctx context.Context, key string, opts ...ObjectOption) ([]byte, error) {
	body, _, err := s.DownloadStream(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read s3 object failed: %w", err)
	}

	return data, nil
}

func (s *Service) DownloadStream(ctx context.Context, key string, opts ...ObjectOption) (io.ReadCloser, *ObjectInfo, error) {
	service, err := ensureService(s)
	if err != nil {
		return nil, nil, err
	}

	objectOpts, bucket, err := service.resolveObjectOptions(opts...)
	if err != nil {
		return nil, nil, err
	}

	input, err := buildGetInput(bucket, key, objectOpts)
	if err != nil {
		return nil, nil, err
	}

	output, err := service.client.GetObject(ctxOrBackground(ctx), input)
	if err != nil {
		return nil, nil, classifyS3Error(err)
	}

	return output.Body, mapGetObject(bucket, key, output), nil
}

func (s *Service) UploadBytes(ctx context.Context, key string, data []byte, opts ...PutOption) (*UploadResult, error) {
	return s.UploadReader(ctx, key, bytes.NewReader(data), opts...)
}

func (s *Service) UploadReader(ctx context.Context, key string, body io.Reader, opts ...PutOption) (*UploadResult, error) {
	service, err := ensureService(s)
	if err != nil {
		return nil, err
	}
	if body == nil {
		return nil, errors.New("s3 upload body is required")
	}

	putOpts, bucket, err := service.resolvePutOptions(opts...)
	if err != nil {
		return nil, err
	}

	input, err := buildPutInput(bucket, key, body, putOpts)
	if err != nil {
		return nil, err
	}

	output, err := service.client.PutObject(ctxOrBackground(ctx), input)
	if err != nil {
		return nil, classifyS3Error(err)
	}

	return mapPutObject(bucket, key, output), nil
}

func (s *Service) UploadFile(ctx context.Context, key string, path string, opts ...PutOption) (*UploadResult, error) {
	service, err := ensureService(s)
	if err != nil {
		return nil, err
	}

	filePath := strings.TrimSpace(path)
	if filePath == "" {
		return nil, errors.New("s3 upload file path is required")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open upload file failed: %w", err)
	}
	defer file.Close()

	return service.UploadReader(ctx, key, file, opts...)
}

func (s *Service) PresignGet(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureService(s)
	if err != nil {
		return "", nil, err
	}

	presignOpts, bucket, err := service.resolvePresignOptions(opts...)
	if err != nil {
		return "", nil, err
	}

	input, err := buildPresignGetInput(bucket, key, presignOpts)
	if err != nil {
		return "", nil, err
	}

	request, err := service.signer.PresignGetObject(ctxOrBackground(ctx), input, presignExpireFn(presignOpts.Expire))
	if err != nil {
		return "", nil, classifyS3Error(err)
	}

	return request.URL, cloneHTTPHeader(request.SignedHeader), nil
}

func (s *Service) PresignPut(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureService(s)
	if err != nil {
		return "", nil, err
	}

	presignOpts, bucket, err := service.resolvePresignOptions(opts...)
	if err != nil {
		return "", nil, err
	}

	input, err := buildPresignPutInput(bucket, key, presignOpts)
	if err != nil {
		return "", nil, err
	}

	request, err := service.signer.PresignPutObject(ctxOrBackground(ctx), input, presignExpireFn(presignOpts.Expire))
	if err != nil {
		return "", nil, classifyS3Error(err)
	}

	return request.URL, cloneHTTPHeader(request.SignedHeader), nil
}

func (s *Service) PresignDelete(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureService(s)
	if err != nil {
		return "", nil, err
	}

	presignOpts, bucket, err := service.resolvePresignOptions(opts...)
	if err != nil {
		return "", nil, err
	}

	input, err := buildPresignDeleteInput(bucket, key, presignOpts)
	if err != nil {
		return "", nil, err
	}

	request, err := service.signer.PresignDeleteObject(ctxOrBackground(ctx), input, presignExpireFn(presignOpts.Expire))
	if err != nil {
		return "", nil, classifyS3Error(err)
	}

	return request.URL, cloneHTTPHeader(request.SignedHeader), nil
}

func (b *Bucket) RawName() string {
	if b == nil {
		return ""
	}
	return b.name
}

func (b *Bucket) HeadObject(ctx context.Context, key string, opts ...ObjectOption) (*ObjectInfo, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, err
	}
	return service.HeadObject(ctx, key, prependObjectBucket(b.name, opts)...)
}

func (b *Bucket) DeleteObject(ctx context.Context, key string, opts ...ObjectOption) error {
	service, err := ensureBucketService(b)
	if err != nil {
		return err
	}
	return service.DeleteObject(ctx, key, prependObjectBucket(b.name, opts)...)
}

func (b *Bucket) DownloadBytes(ctx context.Context, key string, opts ...ObjectOption) ([]byte, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, err
	}
	return service.DownloadBytes(ctx, key, prependObjectBucket(b.name, opts)...)
}

func (b *Bucket) DownloadStream(ctx context.Context, key string, opts ...ObjectOption) (io.ReadCloser, *ObjectInfo, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, nil, err
	}
	return service.DownloadStream(ctx, key, prependObjectBucket(b.name, opts)...)
}

func (b *Bucket) UploadBytes(ctx context.Context, key string, data []byte, opts ...PutOption) (*UploadResult, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, err
	}
	return service.UploadBytes(ctx, key, data, prependPutBucket(b.name, opts)...)
}

func (b *Bucket) UploadReader(ctx context.Context, key string, body io.Reader, opts ...PutOption) (*UploadResult, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, err
	}
	return service.UploadReader(ctx, key, body, prependPutBucket(b.name, opts)...)
}

func (b *Bucket) UploadFile(ctx context.Context, key string, path string, opts ...PutOption) (*UploadResult, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return nil, err
	}
	return service.UploadFile(ctx, key, path, prependPutBucket(b.name, opts)...)
}

func (b *Bucket) PresignGet(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return "", nil, err
	}
	return service.PresignGet(ctx, key, prependPresignBucket(b.name, opts)...)
}

func (b *Bucket) PresignPut(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return "", nil, err
	}
	return service.PresignPut(ctx, key, prependPresignBucket(b.name, opts)...)
}

func (b *Bucket) PresignDelete(ctx context.Context, key string, opts ...PresignOption) (string, http.Header, error) {
	service, err := ensureBucketService(b)
	if err != nil {
		return "", nil, err
	}
	return service.PresignDelete(ctx, key, prependPresignBucket(b.name, opts)...)
}

func WithBucket(name string) bucketOption {
	return bucketOption{name: strings.TrimSpace(name)}
}

func WithContentType(value string) contentTypeOption {
	return contentTypeOption{value: strings.TrimSpace(value)}
}

func WithCacheControl(value string) cacheControlOption {
	return cacheControlOption{value: strings.TrimSpace(value)}
}

func WithContentDisposition(value string) contentDispositionOption {
	return contentDispositionOption{value: strings.TrimSpace(value)}
}

func WithMetadata(values map[string]string) metadataOption {
	return metadataOption{values: cloneStringMap(values)}
}

func WithExpire(expire time.Duration) expireOption {
	return expireOption{value: expire}
}

func (o bucketOption) applyObject(opts *ObjectOptions) {
	opts.Bucket = o.name
}

func (o bucketOption) applyPut(opts *PutOptions) {
	opts.Bucket = o.name
}

func (o bucketOption) applyPresign(opts *PresignOptions) {
	opts.Bucket = o.name
}

func (o contentTypeOption) applyPut(opts *PutOptions) {
	opts.ContentType = o.value
}

func (o contentTypeOption) applyPresign(opts *PresignOptions) {
	opts.ContentType = o.value
}

func (o cacheControlOption) applyPut(opts *PutOptions) {
	opts.CacheControl = o.value
}

func (o cacheControlOption) applyPresign(opts *PresignOptions) {
	opts.CacheControl = o.value
}

func (o contentDispositionOption) applyPut(opts *PutOptions) {
	opts.ContentDisposition = o.value
}

func (o contentDispositionOption) applyPresign(opts *PresignOptions) {
	opts.ContentDisposition = o.value
}

func (o metadataOption) applyPut(opts *PutOptions) {
	opts.Metadata = cloneStringMap(o.values)
}

func (o metadataOption) applyPresign(opts *PresignOptions) {
	opts.Metadata = cloneStringMap(o.values)
}

func (o expireOption) applyPresign(opts *PresignOptions) {
	opts.Expire = o.value
}

func normalizeConfig(cfg Config) (Config, error) {
	normalized := cfg
	normalized.Region = strings.TrimSpace(normalized.Region)
	normalized.Endpoint = strings.TrimSpace(normalized.Endpoint)
	normalized.AccessKey = strings.TrimSpace(normalized.AccessKey)
	normalized.SecretKey = strings.TrimSpace(normalized.SecretKey)
	normalized.SessionToken = strings.TrimSpace(normalized.SessionToken)
	normalized.Bucket = strings.TrimSpace(normalized.Bucket)

	if normalized.Region == "" {
		return Config{}, errors.New("s3 region is required")
	}
	if (normalized.AccessKey == "") != (normalized.SecretKey == "") {
		return Config{}, errors.New("s3 access key and secret key must be provided together")
	}
	if normalized.SessionToken != "" && normalized.AccessKey == "" {
		return Config{}, errors.New("s3 session token requires access key and secret key")
	}
	if normalized.PresignExpire <= 0 {
		normalized.PresignExpire = defaultPresignExpire
	}
	if normalized.Endpoint != "" {
		endpoint, err := normalizeEndpoint(normalized.Endpoint, normalized.UseSSL)
		if err != nil {
			return Config{}, err
		}
		normalized.Endpoint = endpoint
	}

	return normalized, nil
}

func buildAWSConfig(ctx context.Context, cfg Config) (aws.Config, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			cfg.SessionToken,
		)))
	} else {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}))
	}

	awsCfg, err := loadAWSConfig(ctx, loadOptions...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load aws config failed: %w", err)
	}

	return awsCfg, nil
}

func buildClientOptionFns(cfg Config) []func(*awss3.Options) {
	optionFns := make([]func(*awss3.Options), 0, 2)

	if cfg.Endpoint != "" {
		optionFns = append(optionFns, func(opts *awss3.Options) {
			opts.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	optionFns = append(optionFns, func(opts *awss3.Options) {
		opts.UsePathStyle = cfg.ForcePathStyle
	})

	return optionFns
}

func normalizeEndpoint(endpoint string, useSSL bool) (string, error) {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return "", errors.New("s3 endpoint is required")
	}

	if !strings.Contains(trimmed, "://") {
		scheme := "https"
		if !useSSL {
			scheme = "http"
		}
		trimmed = scheme + "://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid s3 endpoint: %w", err)
	}
	if parsed.Host == "" {
		return "", errors.New("invalid s3 endpoint: host is required")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

func (s *Service) resolveObjectOptions(opts ...ObjectOption) (ObjectOptions, string, error) {
	objectOpts := ObjectOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt.applyObject(&objectOpts)
		}
	}

	bucket := strings.TrimSpace(objectOpts.Bucket)
	if bucket == "" {
		bucket = strings.TrimSpace(s.bucket)
	}
	if bucket == "" {
		return ObjectOptions{}, "", errors.New("s3 bucket is required")
	}

	return objectOpts, bucket, nil
}

func (s *Service) resolvePutOptions(opts ...PutOption) (PutOptions, string, error) {
	putOpts := PutOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt.applyPut(&putOpts)
		}
	}

	bucket := strings.TrimSpace(putOpts.Bucket)
	if bucket == "" {
		bucket = strings.TrimSpace(s.bucket)
	}
	if bucket == "" {
		return PutOptions{}, "", errors.New("s3 bucket is required")
	}

	putOpts.Metadata = cloneStringMap(putOpts.Metadata)
	return putOpts, bucket, nil
}

func (s *Service) resolvePresignOptions(opts ...PresignOption) (PresignOptions, string, error) {
	presignOpts := PresignOptions{Expire: s.cfg.PresignExpire}
	for _, opt := range opts {
		if opt != nil {
			opt.applyPresign(&presignOpts)
		}
	}

	if presignOpts.Expire <= 0 {
		presignOpts.Expire = s.cfg.PresignExpire
		if presignOpts.Expire <= 0 {
			presignOpts.Expire = defaultPresignExpire
		}
	}

	bucket := strings.TrimSpace(presignOpts.Bucket)
	if bucket == "" {
		bucket = strings.TrimSpace(s.bucket)
	}
	if bucket == "" {
		return PresignOptions{}, "", errors.New("s3 bucket is required")
	}

	presignOpts.Metadata = cloneStringMap(presignOpts.Metadata)
	return presignOpts, bucket, nil
}

func buildHeadInput(bucket, key string, _ ObjectOptions) (*awss3.HeadObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, nil
}

func buildDeleteInput(bucket, key string, _ ObjectOptions) (*awss3.DeleteObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, nil
}

func buildGetInput(bucket, key string, _ ObjectOptions) (*awss3.GetObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, nil
}

func buildPutInput(bucket, key string, body io.Reader, opts PutOptions) (*awss3.PutObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(key),
		Body:               body,
		ContentType:        awsStringOrNil(opts.ContentType),
		CacheControl:       awsStringOrNil(opts.CacheControl),
		ContentDisposition: awsStringOrNil(opts.ContentDisposition),
		Metadata:           cloneStringMap(opts.Metadata),
	}, nil
}

func buildPresignGetInput(bucket, key string, opts PresignOptions) (*awss3.GetObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	input := &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if value := strings.TrimSpace(opts.CacheControl); value != "" {
		input.ResponseCacheControl = aws.String(value)
	}
	if value := strings.TrimSpace(opts.ContentDisposition); value != "" {
		input.ResponseContentDisposition = aws.String(value)
	}
	if value := strings.TrimSpace(opts.ContentType); value != "" {
		input.ResponseContentType = aws.String(value)
	}

	return input, nil
}

func buildPresignDeleteInput(bucket, key string, _ PresignOptions) (*awss3.DeleteObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, nil
}

func buildPresignPutInput(bucket, key string, opts PresignOptions) (*awss3.PutObjectInput, error) {
	if key == "" {
		return nil, errors.New("s3 object key is required")
	}

	return &awss3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(key),
		ContentType:        awsStringOrNil(opts.ContentType),
		CacheControl:       awsStringOrNil(opts.CacheControl),
		ContentDisposition: awsStringOrNil(opts.ContentDisposition),
		Metadata:           cloneStringMap(opts.Metadata),
	}, nil
}

func classifyS3Error(err error) error {
	if err == nil {
		return nil
	}

	var responseErr *smithyhttp.ResponseError
	if errors.As(err, &responseErr) && responseErr.HTTPStatusCode() == http.StatusNotFound {
		return ErrObjectNotFound
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "NoSuchBucket", "404":
			return ErrObjectNotFound
		}
	}

	return err
}

func mapHeadObject(bucket, key string, output *awss3.HeadObjectOutput) *ObjectInfo {
	if output == nil {
		return &ObjectInfo{Bucket: bucket, Key: key}
	}

	return &ObjectInfo{
		Bucket:        bucket,
		Key:           key,
		ETag:          aws.ToString(output.ETag),
		ContentType:   aws.ToString(output.ContentType),
		ContentLength: aws.ToInt64(output.ContentLength),
		LastModified:  aws.ToTime(output.LastModified),
		Metadata:      cloneStringMap(output.Metadata),
	}
}

func mapGetObject(bucket, key string, output *awss3.GetObjectOutput) *ObjectInfo {
	if output == nil {
		return &ObjectInfo{Bucket: bucket, Key: key}
	}

	return &ObjectInfo{
		Bucket:        bucket,
		Key:           key,
		ETag:          aws.ToString(output.ETag),
		ContentType:   aws.ToString(output.ContentType),
		ContentLength: aws.ToInt64(output.ContentLength),
		LastModified:  aws.ToTime(output.LastModified),
		Metadata:      cloneStringMap(output.Metadata),
	}
}

func mapPutObject(bucket, key string, output *awss3.PutObjectOutput) *UploadResult {
	if output == nil {
		return &UploadResult{Bucket: bucket, Key: key}
	}

	return &UploadResult{
		Bucket:    bucket,
		Key:       key,
		ETag:      aws.ToString(output.ETag),
		VersionID: aws.ToString(output.VersionId),
	}
}

func presignExpireFn(expire time.Duration) func(*awss3.PresignOptions) {
	return func(opts *awss3.PresignOptions) {
		opts.Expires = expire
	}
}

func prependObjectBucket(name string, opts []ObjectOption) []ObjectOption {
	return append([]ObjectOption{WithBucket(name)}, opts...)
}

func prependPutBucket(name string, opts []PutOption) []PutOption {
	return append([]PutOption{WithBucket(name)}, opts...)
}

func prependPresignBucket(name string, opts []PresignOption) []PresignOption {
	return append([]PresignOption{WithBucket(name)}, opts...)
}

func ensureService(s *Service) (*Service, error) {
	if s == nil {
		return nil, errors.New("s3 service is nil")
	}
	return s, nil
}

func ensureBucketService(b *Bucket) (*Service, error) {
	if b == nil || b.service == nil {
		return nil, errors.New("s3 bucket service is nil")
	}
	return b.service, nil
}

func awsStringOrNil(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return aws.String(trimmed)
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneHTTPHeader(header http.Header) http.Header {
	if header == nil {
		return nil
	}
	return header.Clone()
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

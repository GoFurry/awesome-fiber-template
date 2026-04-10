package mongodb

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Config struct {
	URI                    string
	Hosts                  []string
	Username               string
	Password               string
	AuthSource             string
	Database               string
	ReplicaSet             string
	Direct                 bool
	AppName                string
	MinPoolSize            uint64
	MaxPoolSize            uint64
	MaxConnIdleTime        time.Duration
	ConnectTimeout         time.Duration
	SocketTimeout          time.Duration
	ServerSelectionTimeout time.Duration
	RetryReads             *bool
	RetryWrites            *bool
	TLS                    *TLSConfig
}

type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
}

type Service struct {
	cfg      Config
	raw      *mongo.Client
	client   clientAPI
	database string
}

type Collection struct {
	raw        *mongo.Collection
	collection collectionAPI
	database   string
	name       string
	err        error
}

type clientAPI interface {
	Database(name string, opts ...options.Lister[options.DatabaseOptions]) databaseAPI
	Ping(ctx context.Context, rp *readpref.ReadPref) error
	Disconnect(ctx context.Context) error
}

type databaseAPI interface {
	Collection(name string, opts ...options.Lister[options.CollectionOptions]) collectionAPI
}

type collectionAPI interface {
	InsertOne(ctx context.Context, doc any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error)
	InsertMany(ctx context.Context, docs []any, opts ...options.Lister[options.InsertManyOptions]) (*mongo.InsertManyResult, error)
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) singleResultAPI
	Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (cursorAPI, error)
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error)
	ReplaceOne(ctx context.Context, filter any, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error)
	CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error)
}

type singleResultAPI interface {
	Decode(v any) error
	Err() error
}

type cursorAPI interface {
	All(ctx context.Context, results any) error
	Close(ctx context.Context) error
}

type mongoClientAdapter struct {
	raw *mongo.Client
}

type mongoDatabaseAdapter struct {
	raw *mongo.Database
}

type mongoCollectionAdapter struct {
	raw *mongo.Collection
}

type mongoSingleResultAdapter struct {
	raw *mongo.SingleResult
}

type mongoCursorAdapter struct {
	raw *mongo.Cursor
}

var connectMongo = func(opts ...*options.ClientOptions) (*mongo.Client, error) {
	return mongo.Connect(opts...)
}

var wrapMongoClient = func(raw *mongo.Client) clientAPI {
	return mongoClientAdapter{raw: raw}
}

func New(ctx context.Context, cfg Config) (*Service, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	clientOptions, err := buildClientOptions(normalized)
	if err != nil {
		return nil, err
	}

	raw, err := connectMongo(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb failed: %w", err)
	}

	service := &Service{
		cfg:      normalized,
		raw:      raw,
		client:   wrapMongoClient(raw),
		database: normalized.Database,
	}

	if ctx == nil {
		ctx = context.Background()
	}
	if err := service.Ping(ctx); err != nil {
		_ = service.Close(ctx)
		return nil, fmt.Errorf("ping mongodb failed: %w", err)
	}

	return service, nil
}

func (s *Service) Client() *mongo.Client {
	if s == nil {
		return nil
	}
	return s.raw
}

func (s *Service) Database(name ...string) *mongo.Database {
	if s == nil || s.raw == nil {
		return nil
	}

	databaseName := s.resolveDatabaseName(name...)
	if databaseName == "" {
		return nil
	}

	return s.raw.Database(databaseName)
}

func (s *Service) Collection(name string) *Collection {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return &Collection{err: errors.New("collection name is required")}
	}

	databaseName := s.resolveDatabaseName()
	if databaseName == "" {
		return &Collection{
			name: trimmedName,
			err:  errors.New("mongodb database is not configured"),
		}
	}
	if s == nil || s.client == nil {
		return &Collection{
			name:     trimmedName,
			database: databaseName,
			err:      errors.New("mongodb client is not initialized"),
		}
	}

	var rawCollection *mongo.Collection
	if s.raw != nil {
		rawDatabase := s.raw.Database(databaseName)
		if rawDatabase != nil {
			rawCollection = rawDatabase.Collection(trimmedName)
		}
	}

	return &Collection{
		raw:        rawCollection,
		collection: s.client.Database(databaseName).Collection(trimmedName),
		database:   databaseName,
		name:       trimmedName,
	}
}

func (s *Service) Ping(ctx context.Context) error {
	if s == nil || s.client == nil {
		return errors.New("mongodb client is not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.client.Ping(ctx, readpref.Primary())
}

func (s *Service) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.client.Disconnect(ctx)
}

func (c *Collection) Raw() *mongo.Collection {
	if c == nil {
		return nil
	}
	return c.raw
}

func (c *Collection) InsertOne(ctx context.Context, doc any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.InsertOne(ctx, doc, opts...)
}

func (c *Collection) InsertMany(ctx context.Context, docs []any, opts ...options.Lister[options.InsertManyOptions]) (*mongo.InsertManyResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.InsertMany(ctx, docs, opts...)
}

func (c *Collection) FindOne(ctx context.Context, filter any, result any, opts ...options.Lister[options.FindOneOptions]) error {
	if err := c.ensureReady(); err != nil {
		return err
	}
	if err := validateFindOneTarget(result); err != nil {
		return err
	}
	return c.collection.FindOne(ctx, filter, opts...).Decode(result)
}

func (c *Collection) FindMany(ctx context.Context, filter any, result any, opts ...options.Lister[options.FindOptions]) error {
	if err := c.ensureReady(); err != nil {
		return err
	}
	if err := validateFindManyTarget(result); err != nil {
		return err
	}

	cursor, err := c.collection.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	return cursor.All(ctx, result)
}

func (c *Collection) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.UpdateOne(ctx, filter, update, opts...)
}

func (c *Collection) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.UpdateMany(ctx, filter, update, opts...)
}

func (c *Collection) ReplaceOne(ctx context.Context, filter any, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*mongo.UpdateResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.ReplaceOne(ctx, filter, replacement, opts...)
}

func (c *Collection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.DeleteOne(ctx, filter, opts...)
}

func (c *Collection) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.collection.DeleteMany(ctx, filter, opts...)
}

func (c *Collection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if err := c.ensureReady(); err != nil {
		return 0, err
	}
	return c.collection.CountDocuments(ctx, filter, opts...)
}

func (s *Service) resolveDatabaseName(name ...string) string {
	if len(name) > 0 {
		explicit := strings.TrimSpace(name[0])
		if explicit != "" {
			return explicit
		}
		return ""
	}

	return strings.TrimSpace(s.database)
}

func (c *Collection) ensureReady() error {
	if c == nil {
		return errors.New("mongodb collection is nil")
	}
	if c.err != nil {
		return c.err
	}
	if strings.TrimSpace(c.name) == "" {
		return errors.New("collection name is required")
	}
	if strings.TrimSpace(c.database) == "" {
		return errors.New("mongodb database is not configured")
	}
	if c.collection == nil {
		return errors.New("mongodb collection is not initialized")
	}
	return nil
}

func validateFindOneTarget(result any) error {
	if result == nil {
		return errors.New("find one result target must be a non-nil pointer")
	}
	value := reflect.ValueOf(result)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("find one result target must be a non-nil pointer")
	}
	return nil
}

func validateFindManyTarget(result any) error {
	if result == nil {
		return errors.New("find many result target must be a non-nil slice pointer")
	}
	value := reflect.ValueOf(result)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("find many result target must be a non-nil slice pointer")
	}
	if value.Elem().Kind() != reflect.Slice {
		return errors.New("find many result target must be a slice pointer")
	}
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	normalized := cfg
	normalized.URI = strings.TrimSpace(normalized.URI)
	normalized.Database = strings.TrimSpace(normalized.Database)
	normalized.Username = strings.TrimSpace(normalized.Username)
	normalized.Password = strings.TrimSpace(normalized.Password)
	normalized.AuthSource = strings.TrimSpace(normalized.AuthSource)
	normalized.ReplicaSet = strings.TrimSpace(normalized.ReplicaSet)
	normalized.AppName = strings.TrimSpace(normalized.AppName)

	if normalized.URI == "" {
		hosts := make([]string, 0, len(normalized.Hosts))
		for _, host := range normalized.Hosts {
			trimmed := strings.TrimSpace(host)
			if trimmed == "" {
				continue
			}
			hosts = append(hosts, trimmed)
		}
		if len(hosts) == 0 {
			return Config{}, errors.New("mongodb uri or hosts is required")
		}
		normalized.Hosts = hosts
	}

	return normalized, nil
}

func buildClientOptions(cfg Config) (*options.ClientOptions, error) {
	clientOptions := options.Client()

	if cfg.URI != "" {
		clientOptions.ApplyURI(cfg.URI)
	} else {
		clientOptions.ApplyURI(buildMongoURI(cfg))
	}

	if cfg.URI == "" && len(cfg.Hosts) > 0 {
		clientOptions.SetHosts(cfg.Hosts)
	}
	if cfg.Username != "" || cfg.Password != "" || cfg.AuthSource != "" {
		clientOptions.SetAuth(options.Credential{
			Username:   cfg.Username,
			Password:   cfg.Password,
			AuthSource: cfg.AuthSource,
		})
	}
	if cfg.ReplicaSet != "" {
		clientOptions.SetReplicaSet(cfg.ReplicaSet)
	}
	if cfg.Direct {
		clientOptions.SetDirect(true)
	}
	if cfg.AppName != "" {
		clientOptions.SetAppName(cfg.AppName)
	}
	if cfg.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MaxConnIdleTime > 0 {
		clientOptions.SetMaxConnIdleTime(cfg.MaxConnIdleTime)
	}
	if cfg.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(cfg.ConnectTimeout)
	}
	if cfg.ServerSelectionTimeout > 0 {
		clientOptions.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
	}
	if cfg.RetryReads != nil {
		clientOptions.SetRetryReads(*cfg.RetryReads)
	}
	if cfg.RetryWrites != nil {
		clientOptions.SetRetryWrites(*cfg.RetryWrites)
	}
	if cfg.TLS != nil && cfg.TLS.Enabled {
		clientOptions.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		})
	}

	if err := clientOptions.Validate(); err != nil {
		return nil, fmt.Errorf("mongodb client options are invalid: %w", err)
	}
	return clientOptions, nil
}

func buildMongoURI(cfg Config) string {
	query := url.Values{}
	if cfg.AuthSource != "" {
		query.Set("authSource", cfg.AuthSource)
	}
	if cfg.ReplicaSet != "" {
		query.Set("replicaSet", cfg.ReplicaSet)
	}
	if cfg.Direct {
		query.Set("directConnection", "true")
	}
	if cfg.ConnectTimeout > 0 {
		query.Set("connectTimeoutMS", strconvMillis(cfg.ConnectTimeout))
	}
	if cfg.SocketTimeout > 0 {
		query.Set("socketTimeoutMS", strconvMillis(cfg.SocketTimeout))
	}
	if cfg.ServerSelectionTimeout > 0 {
		query.Set("serverSelectionTimeoutMS", strconvMillis(cfg.ServerSelectionTimeout))
	}
	if cfg.RetryReads != nil {
		query.Set("retryReads", boolToString(*cfg.RetryReads))
	}
	if cfg.RetryWrites != nil {
		query.Set("retryWrites", boolToString(*cfg.RetryWrites))
	}
	if cfg.TLS != nil && cfg.TLS.Enabled {
		query.Set("tls", "true")
	}

	var builder strings.Builder
	builder.WriteString("mongodb://")
	if cfg.Username != "" || cfg.Password != "" {
		builder.WriteString(url.UserPassword(cfg.Username, cfg.Password).String())
		builder.WriteString("@")
	}
	builder.WriteString(strings.Join(cfg.Hosts, ","))

	if cfg.Database != "" || len(query) > 0 {
		builder.WriteString("/")
	}
	if cfg.Database != "" {
		builder.WriteString(url.PathEscape(cfg.Database))
	}
	if len(query) > 0 {
		builder.WriteString("?")
		builder.WriteString(query.Encode())
	}

	return builder.String()
}

func strconvMillis(duration time.Duration) string {
	return fmt.Sprintf("%d", duration.Milliseconds())
}

func boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (m mongoClientAdapter) Database(name string, opts ...options.Lister[options.DatabaseOptions]) databaseAPI {
	return mongoDatabaseAdapter{raw: m.raw.Database(name, opts...)}
}

func (m mongoClientAdapter) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return m.raw.Ping(ctx, rp)
}

func (m mongoClientAdapter) Disconnect(ctx context.Context) error {
	return m.raw.Disconnect(ctx)
}

func (m mongoDatabaseAdapter) Collection(name string, opts ...options.Lister[options.CollectionOptions]) collectionAPI {
	return mongoCollectionAdapter{raw: m.raw.Collection(name, opts...)}
}

func (m mongoCollectionAdapter) InsertOne(ctx context.Context, doc any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	return m.raw.InsertOne(ctx, doc, opts...)
}

func (m mongoCollectionAdapter) InsertMany(ctx context.Context, docs []any, opts ...options.Lister[options.InsertManyOptions]) (*mongo.InsertManyResult, error) {
	return m.raw.InsertMany(ctx, docs, opts...)
}

func (m mongoCollectionAdapter) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) singleResultAPI {
	return mongoSingleResultAdapter{raw: m.raw.FindOne(ctx, filter, opts...)}
}

func (m mongoCollectionAdapter) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (cursorAPI, error) {
	cursor, err := m.raw.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return mongoCursorAdapter{raw: cursor}, nil
}

func (m mongoCollectionAdapter) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return m.raw.UpdateOne(ctx, filter, update, opts...)
}

func (m mongoCollectionAdapter) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	return m.raw.UpdateMany(ctx, filter, update, opts...)
}

func (m mongoCollectionAdapter) ReplaceOne(ctx context.Context, filter any, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*mongo.UpdateResult, error) {
	return m.raw.ReplaceOne(ctx, filter, replacement, opts...)
}

func (m mongoCollectionAdapter) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	return m.raw.DeleteOne(ctx, filter, opts...)
}

func (m mongoCollectionAdapter) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	return m.raw.DeleteMany(ctx, filter, opts...)
}

func (m mongoCollectionAdapter) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return m.raw.CountDocuments(ctx, filter, opts...)
}

func (m mongoSingleResultAdapter) Decode(v any) error {
	return m.raw.Decode(v)
}

func (m mongoSingleResultAdapter) Err() error {
	return m.raw.Err()
}

func (m mongoCursorAdapter) All(ctx context.Context, results any) error {
	return m.raw.All(ctx, results)
}

func (m mongoCursorAdapter) Close(ctx context.Context) error {
	return m.raw.Close(ctx)
}

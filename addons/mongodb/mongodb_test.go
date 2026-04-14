package mongodb

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func TestNormalizeConfigRequiresURIOrHosts(t *testing.T) {
	t.Parallel()

	_, err := normalizeConfig(Config{})
	if err == nil {
		t.Fatalf("expected config validation error")
	}
}

func TestBuildClientOptionsDSNPriority(t *testing.T) {
	t.Parallel()

	opts, err := buildClientOptions(Config{
		URI:      "mongodb://uri-host:27017/uri_db?retryWrites=false",
		Hosts:    []string{"structured-host:27017"},
		Database: "structured_db",
		Direct:   true,
	})
	if err != nil {
		t.Fatalf("build client options failed: %v", err)
	}

	if len(opts.Hosts) != 1 || opts.Hosts[0] != "uri-host:27017" {
		t.Fatalf("expected URI hosts to win, got %+v", opts.Hosts)
	}
	if opts.Auth != nil {
		t.Fatalf("expected URI auth to remain untouched, got %+v", opts.Auth)
	}
}

func TestBuildClientOptionsURIAuthPriority(t *testing.T) {
	t.Parallel()

	opts, err := buildClientOptions(Config{
		URI:        "mongodb://uri-user:uri-secret@uri-host:27017/uri_db?authSource=uri_admin",
		Username:   "structured-user",
		Password:   "structured-secret",
		AuthSource: "structured_admin",
	})
	if err != nil {
		t.Fatalf("build client options failed: %v", err)
	}

	if opts.Auth == nil || opts.Auth.Username != "uri-user" || opts.Auth.Password != "uri-secret" {
		t.Fatalf("expected URI auth to win, got %+v", opts.Auth)
	}
}

func TestBuildClientOptionsStructuredConfig(t *testing.T) {
	t.Parallel()

	retryReads := true
	retryWrites := false
	opts, err := buildClientOptions(Config{
		Hosts:                  []string{"mongo-1:27017", "mongo-2:27017"},
		Username:               "root",
		Password:               "secret",
		AuthSource:             "admin",
		Database:               "app",
		ReplicaSet:             "rs0",
		Direct:                 false,
		AppName:                "awesome-fiber-template",
		MinPoolSize:            5,
		MaxPoolSize:            50,
		MaxConnIdleTime:        30 * time.Second,
		ConnectTimeout:         5 * time.Second,
		SocketTimeout:          9 * time.Second,
		ServerSelectionTimeout: 15 * time.Second,
		RetryReads:             &retryReads,
		RetryWrites:            &retryWrites,
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		t.Fatalf("build client options failed: %v", err)
	}

	if !slices.Equal(opts.Hosts, []string{"mongo-1:27017", "mongo-2:27017"}) {
		t.Fatalf("unexpected hosts: %+v", opts.Hosts)
	}
	if opts.Auth == nil || opts.Auth.Username != "root" || opts.Auth.AuthSource != "admin" {
		t.Fatalf("unexpected auth config: %+v", opts.Auth)
	}
	if opts.ReplicaSet == nil || *opts.ReplicaSet != "rs0" {
		t.Fatalf("unexpected replica set: %+v", opts.ReplicaSet)
	}
	if opts.AppName == nil || *opts.AppName != "awesome-fiber-template" {
		t.Fatalf("unexpected app name: %+v", opts.AppName)
	}
	if opts.MinPoolSize == nil || *opts.MinPoolSize != 5 {
		t.Fatalf("unexpected min pool size: %+v", opts.MinPoolSize)
	}
	if opts.MaxPoolSize == nil || *opts.MaxPoolSize != 50 {
		t.Fatalf("unexpected max pool size: %+v", opts.MaxPoolSize)
	}
	if opts.MaxConnIdleTime == nil || *opts.MaxConnIdleTime != 30*time.Second {
		t.Fatalf("unexpected idle timeout: %+v", opts.MaxConnIdleTime)
	}
	if opts.ConnectTimeout == nil || *opts.ConnectTimeout != 5*time.Second {
		t.Fatalf("unexpected connect timeout: %+v", opts.ConnectTimeout)
	}
	if opts.ServerSelectionTimeout == nil || *opts.ServerSelectionTimeout != 15*time.Second {
		t.Fatalf("unexpected server selection timeout: %+v", opts.ServerSelectionTimeout)
	}
	if opts.RetryReads == nil || *opts.RetryReads != true {
		t.Fatalf("unexpected retry reads: %+v", opts.RetryReads)
	}
	if opts.RetryWrites == nil || *opts.RetryWrites != false {
		t.Fatalf("unexpected retry writes: %+v", opts.RetryWrites)
	}
	if opts.TLSConfig == nil || !opts.TLSConfig.InsecureSkipVerify {
		t.Fatalf("unexpected tls config: %+v", opts.TLSConfig)
	}

	uri := buildMongoURI(Config{
		Hosts:                  []string{"mongo-1:27017", "mongo-2:27017"},
		Username:               "root",
		Password:               "secret",
		AuthSource:             "admin",
		Database:               "app",
		ReplicaSet:             "rs0",
		Direct:                 true,
		ConnectTimeout:         5 * time.Second,
		SocketTimeout:          9 * time.Second,
		ServerSelectionTimeout: 15 * time.Second,
		RetryReads:             &retryReads,
		RetryWrites:            &retryWrites,
		TLS:                    &TLSConfig{Enabled: true},
	})
	if !strings.Contains(uri, "mongodb://root:secret@mongo-1:27017,mongo-2:27017/app?") {
		t.Fatalf("unexpected mongo uri host/database segment: %s", uri)
	}
	for _, want := range []string{
		"socketTimeoutMS=9000",
		"replicaSet=rs0",
		"directConnection=true",
		"authSource=admin",
	} {
		if !strings.Contains(uri, want) {
			t.Fatalf("expected mongo uri to contain %q, got %s", want, uri)
		}
	}
}

func TestNewLifecycleHooks(t *testing.T) {
	rawClient := &mongo.Client{}
	mockClient := &mockClient{}

	originalConnect := connectMongo
	originalWrap := wrapMongoClient
	t.Cleanup(func() {
		connectMongo = originalConnect
		wrapMongoClient = originalWrap
	})

	connectMongo = func(opts ...*options.ClientOptions) (*mongo.Client, error) {
		return rawClient, nil
	}
	wrapMongoClient = func(raw *mongo.Client) clientAPI {
		if raw != rawClient {
			t.Fatalf("unexpected raw client")
		}
		return mockClient
	}

	service, err := New(context.Background(), Config{
		Hosts:    []string{"mongo-1:27017"},
		Database: "app",
	})
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}
	if service.Client() != rawClient {
		t.Fatalf("unexpected raw client returned")
	}
	if mockClient.pingCount != 1 {
		t.Fatalf("expected ping once, got %d", mockClient.pingCount)
	}
	if err := service.Close(context.Background()); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if mockClient.disconnectCount != 1 {
		t.Fatalf("expected disconnect once, got %d", mockClient.disconnectCount)
	}
}

func TestNewClosesClientOnPingFailure(t *testing.T) {
	rawClient := &mongo.Client{}
	mockClient := &mockClient{pingErr: errors.New("ping failed")}

	originalConnect := connectMongo
	originalWrap := wrapMongoClient
	t.Cleanup(func() {
		connectMongo = originalConnect
		wrapMongoClient = originalWrap
	})

	connectMongo = func(opts ...*options.ClientOptions) (*mongo.Client, error) {
		return rawClient, nil
	}
	wrapMongoClient = func(raw *mongo.Client) clientAPI {
		return mockClient
	}

	_, err := New(context.Background(), Config{
		Hosts: []string{"mongo-1:27017"},
	})
	if err == nil {
		t.Fatalf("expected new service error")
	}
	if mockClient.disconnectCount != 1 {
		t.Fatalf("expected disconnect on ping failure, got %d", mockClient.disconnectCount)
	}
}

func TestServiceDatabaseAndCollectionBehavior(t *testing.T) {
	t.Parallel()

	mockCollection := &mockCollection{}
	mockDatabase := &mockDatabase{collection: mockCollection}
	service := &Service{
		raw:      &mongo.Client{},
		client:   &mockClient{database: mockDatabase},
		database: "app",
	}

	if service.Database() == nil {
		t.Fatalf("expected default database")
	}
	if service.Database("other") == nil {
		t.Fatalf("expected explicit database")
	}

	collection := service.Collection("users")
	if collection == nil || collection.Raw() == nil {
		t.Fatalf("expected raw collection wrapper")
	}
	if _, err := collection.InsertOne(context.Background(), map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("insert one failed: %v", err)
	}
	if !slices.Equal(mockDatabase.collectionNames, []string{"users"}) {
		t.Fatalf("unexpected collection names: %+v", mockDatabase.collectionNames)
	}

	withoutDefaultDatabase := &Service{
		raw:    &mongo.Client{},
		client: &mockClient{database: mockDatabase},
	}
	if withoutDefaultDatabase.Database() != nil {
		t.Fatalf("expected nil database without configured default")
	}
	err := withoutDefaultDatabase.Collection("users").FindOne(context.Background(), map[string]any{}, &map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "database") {
		t.Fatalf("expected missing database error, got %v", err)
	}
}

func TestCollectionCRUDHelpers(t *testing.T) {
	t.Parallel()

	mockCollection := &mockCollection{
		insertOneResult:  &mongo.InsertOneResult{InsertedID: "1"},
		insertManyResult: &mongo.InsertManyResult{InsertedIDs: []any{"1", "2"}},
		updateResult:     &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1},
		deleteResult:     &mongo.DeleteResult{DeletedCount: 2},
		countResult:      12,
		findOneResult: &mockSingleResult{
			decodeFn: func(target any) error {
				target.(*testDocument).Name = "alice"
				return nil
			},
		},
		findCursor: &mockCursor{
			allFn: func(target any) error {
				docs := target.(*[]testDocument)
				*docs = []testDocument{{Name: "alice"}, {Name: "bob"}}
				return nil
			},
		},
	}

	collection := &Collection{
		raw:        &mongo.Collection{},
		collection: mockCollection,
		database:   "app",
		name:       "users",
	}

	insertOneResult, err := collection.InsertOne(context.Background(), map[string]any{"name": "alice"}, options.InsertOne())
	if err != nil || insertOneResult.InsertedID != "1" {
		t.Fatalf("unexpected insert one result: %+v err=%v", insertOneResult, err)
	}
	insertManyResult, err := collection.InsertMany(context.Background(), []any{1, 2}, options.InsertMany())
	if err != nil || len(insertManyResult.InsertedIDs) != 2 {
		t.Fatalf("unexpected insert many result: %+v err=%v", insertManyResult, err)
	}

	var one testDocument
	if err := collection.FindOne(context.Background(), map[string]any{"name": "alice"}, &one, options.FindOne()); err != nil {
		t.Fatalf("find one failed: %v", err)
	}
	if one.Name != "alice" {
		t.Fatalf("unexpected find one decode: %+v", one)
	}

	var many []testDocument
	if err := collection.FindMany(context.Background(), map[string]any{}, &many, options.Find()); err != nil {
		t.Fatalf("find many failed: %v", err)
	}
	if len(many) != 2 {
		t.Fatalf("unexpected find many decode: %+v", many)
	}

	if _, err := collection.UpdateOne(context.Background(), map[string]any{}, map[string]any{"$set": map[string]any{"name": "updated"}}, options.UpdateOne()); err != nil {
		t.Fatalf("update one failed: %v", err)
	}
	if _, err := collection.UpdateMany(context.Background(), map[string]any{}, map[string]any{"$set": map[string]any{"enabled": true}}, options.UpdateMany()); err != nil {
		t.Fatalf("update many failed: %v", err)
	}
	if _, err := collection.ReplaceOne(context.Background(), map[string]any{"_id": 1}, map[string]any{"name": "replaced"}, options.Replace()); err != nil {
		t.Fatalf("replace one failed: %v", err)
	}
	if _, err := collection.DeleteOne(context.Background(), map[string]any{"_id": 1}, options.DeleteOne()); err != nil {
		t.Fatalf("delete one failed: %v", err)
	}
	if _, err := collection.DeleteMany(context.Background(), map[string]any{}, options.DeleteMany()); err != nil {
		t.Fatalf("delete many failed: %v", err)
	}
	count, err := collection.CountDocuments(context.Background(), map[string]any{}, options.Count())
	if err != nil || count != 12 {
		t.Fatalf("unexpected count result: %d err=%v", count, err)
	}

	if mockCollection.insertOneOpts != 1 || mockCollection.insertManyOpts != 1 || mockCollection.findOneOpts != 1 || mockCollection.findOpts != 1 {
		t.Fatalf("expected options passthrough, got %+v", mockCollection)
	}
}

func TestCollectionFindValidation(t *testing.T) {
	t.Parallel()

	collection := &Collection{
		raw:        &mongo.Collection{},
		collection: &mockCollection{},
		database:   "app",
		name:       "users",
	}

	if err := collection.FindOne(context.Background(), map[string]any{}, testDocument{}); err == nil {
		t.Fatalf("expected invalid find one target error")
	}
	if err := collection.FindMany(context.Background(), map[string]any{}, &testDocument{}); err == nil {
		t.Fatalf("expected invalid find many target error")
	}

	expected := mongo.ErrNoDocuments
	collection.collection = &mockCollection{
		findOneResult: &mockSingleResult{
			decodeErr: expected,
		},
	}
	err := collection.FindOne(context.Background(), map[string]any{}, &testDocument{})
	if !errors.Is(err, expected) {
		t.Fatalf("expected ErrNoDocuments, got %v", err)
	}
}

type testDocument struct {
	Name string `bson:"name"`
}

type mockClient struct {
	database        databaseAPI
	pingErr         error
	disconnectErr   error
	pingCount       int
	disconnectCount int
}

func (m *mockClient) Database(name string, opts ...options.Lister[options.DatabaseOptions]) databaseAPI {
	_ = name
	_ = opts
	return m.database
}

func (m *mockClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	_ = ctx
	_ = rp
	m.pingCount++
	return m.pingErr
}

func (m *mockClient) Disconnect(ctx context.Context) error {
	_ = ctx
	m.disconnectCount++
	return m.disconnectErr
}

type mockDatabase struct {
	collection      collectionAPI
	collectionNames []string
}

func (m *mockDatabase) Collection(name string, opts ...options.Lister[options.CollectionOptions]) collectionAPI {
	_ = opts
	m.collectionNames = append(m.collectionNames, name)
	return m.collection
}

type mockCollection struct {
	insertOneResult  *mongo.InsertOneResult
	insertManyResult *mongo.InsertManyResult
	updateResult     *mongo.UpdateResult
	deleteResult     *mongo.DeleteResult
	countResult      int64
	findOneResult    singleResultAPI
	findCursor       cursorAPI

	insertOneOpts  int
	insertManyOpts int
	findOneOpts    int
	findOpts       int
	updateOneOpts  int
	updateManyOpts int
	replaceOneOpts int
	deleteOneOpts  int
	deleteManyOpts int
	countOpts      int
}

func (m *mockCollection) InsertOne(ctx context.Context, doc any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	_ = ctx
	_ = doc
	m.insertOneOpts = len(opts)
	return m.insertOneResult, nil
}

func (m *mockCollection) InsertMany(ctx context.Context, docs []any, opts ...options.Lister[options.InsertManyOptions]) (*mongo.InsertManyResult, error) {
	_ = ctx
	_ = docs
	m.insertManyOpts = len(opts)
	return m.insertManyResult, nil
}

func (m *mockCollection) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) singleResultAPI {
	_ = ctx
	_ = filter
	m.findOneOpts = len(opts)
	return m.findOneResult
}

func (m *mockCollection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (cursorAPI, error) {
	_ = ctx
	_ = filter
	m.findOpts = len(opts)
	return m.findCursor, nil
}

func (m *mockCollection) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	_ = ctx
	_ = filter
	_ = update
	m.updateOneOpts = len(opts)
	return m.updateResult, nil
}

func (m *mockCollection) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	_ = ctx
	_ = filter
	_ = update
	m.updateManyOpts = len(opts)
	return m.updateResult, nil
}

func (m *mockCollection) ReplaceOne(ctx context.Context, filter any, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*mongo.UpdateResult, error) {
	_ = ctx
	_ = filter
	_ = replacement
	m.replaceOneOpts = len(opts)
	return m.updateResult, nil
}

func (m *mockCollection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	_ = ctx
	_ = filter
	m.deleteOneOpts = len(opts)
	return m.deleteResult, nil
}

func (m *mockCollection) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	_ = ctx
	_ = filter
	m.deleteManyOpts = len(opts)
	return m.deleteResult, nil
}

func (m *mockCollection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	_ = ctx
	_ = filter
	m.countOpts = len(opts)
	return m.countResult, nil
}

type mockSingleResult struct {
	decodeFn  func(target any) error
	decodeErr error
	err       error
}

func (m *mockSingleResult) Decode(v any) error {
	if m.decodeFn != nil {
		return m.decodeFn(v)
	}
	return m.decodeErr
}

func (m *mockSingleResult) Err() error {
	return m.err
}

type mockCursor struct {
	allFn    func(target any) error
	allErr   error
	closeErr error
	closed   bool
}

func (m *mockCursor) All(ctx context.Context, results any) error {
	_ = ctx
	if m.allFn != nil {
		return m.allFn(results)
	}
	return m.allErr
}

func (m *mockCursor) Close(ctx context.Context) error {
	_ = ctx
	m.closed = true
	return m.closeErr
}

func TestValidateTargets(t *testing.T) {
	t.Parallel()

	if err := validateFindOneTarget(nil); err == nil {
		t.Fatalf("expected nil pointer error")
	}
	if err := validateFindOneTarget(testDocument{}); err == nil {
		t.Fatalf("expected non-pointer error")
	}
	if err := validateFindManyTarget([]testDocument{}); err == nil {
		t.Fatalf("expected slice pointer error")
	}
	var docs []testDocument
	if err := validateFindManyTarget(&docs); err != nil {
		t.Fatalf("expected valid slice pointer, got %v", err)
	}
}

func TestResolveDatabaseName(t *testing.T) {
	t.Parallel()

	service := &Service{database: "default_db"}
	if got := service.resolveDatabaseName(); got != "default_db" {
		t.Fatalf("unexpected default database: %s", got)
	}
	if got := service.resolveDatabaseName("custom_db"); got != "custom_db" {
		t.Fatalf("unexpected explicit database: %s", got)
	}
	if got := service.resolveDatabaseName(""); got != "" {
		t.Fatalf("expected explicit empty database to stay empty, got %q", got)
	}
}

func TestBuildMongoURIEncodesDatabase(t *testing.T) {
	t.Parallel()

	uri := buildMongoURI(Config{
		Hosts:    []string{"mongo-1:27017"},
		Database: "tenant/app",
	})
	if !strings.Contains(uri, "/tenant%2Fapp") {
		t.Fatalf("expected escaped database path, got %s", uri)
	}
}

func TestCollectionRawNilSafety(t *testing.T) {
	t.Parallel()

	var collection *Collection
	if collection.Raw() != nil {
		t.Fatalf("expected nil raw collection")
	}
}

func TestBooleanStringHelpers(t *testing.T) {
	t.Parallel()

	if boolToString(true) != "true" || boolToString(false) != "false" {
		t.Fatalf("unexpected bool conversion")
	}
	if strconvMillis(1500*time.Millisecond) != "1500" {
		t.Fatalf("unexpected millis conversion")
	}
}

func TestMockFindManyPathUsesCursorClose(t *testing.T) {
	t.Parallel()

	cursor := &mockCursor{
		allFn: func(target any) error {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf([]testDocument{{Name: "alice"}}))
			return nil
		},
	}
	collection := &Collection{
		raw: &mongo.Collection{},
		collection: &mockCollection{
			findCursor: cursor,
		},
		database: "app",
		name:     "users",
	}
	var docs []testDocument
	if err := collection.FindMany(context.Background(), map[string]any{}, &docs); err != nil {
		t.Fatalf("find many failed: %v", err)
	}
	if !cursor.closed {
		t.Fatalf("expected cursor close to be called")
	}
}

func TestNilServiceCollectionIsSafe(t *testing.T) {
	t.Parallel()

	var service *Service
	collection := service.Collection("users")
	if collection == nil {
		t.Fatalf("expected collection wrapper")
	}
	if err := collection.ensureReady(); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected not initialized error, got %v", err)
	}
}

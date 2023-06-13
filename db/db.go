package db

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/nlpodyssey/cybertron/pkg/tasks"
	"github.com/pressly/goose/v3"
	"github.com/schollz/progressbar/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

var (
	//go:embed all:migrations
	migrations     embed.FS
	embeddingModel = "sentence-transformers/all-MiniLM-L6-v2"
	embeddingLimit = 384
)

type Repo struct {
	db *bun.DB
}

func (r *Repo) Close() {
	if err := r.db.Close(); err != nil {
		log.Printf("error: failed to close db: %v\n", err)
	}
}

func New(dbname string) (*Repo, error) {
	enc, err := newEncoder()
	if err != nil {
		return nil, err
	}

	sql.Register("sqlite3_vss", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			err := conn.RegisterFunc("encode_embedding", enc, true)
			return err
		},
	})

	db, err := sql.Open("sqlite3_vss", dbname)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	var version, vector string
	err = db.QueryRow("SELECT vss_version(), vector_to_json(?)", []byte{0x00, 0x00, 0x28, 0x42}).Scan(&version, &vector)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("version=%s vector=%s\n", version, vector)

	r := &Repo{
		db: bun.NewDB(db, sqlitedialect.New()),
	}

	count, _ := r.CountArticles(context.Background())
	if count == 0 {
		log.Println("no articles present, seeding db...")
		if err := seed(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func seed(r *Repo) error {
	defer trackTime("seed complete")

	_, filename, _, _ := runtime.Caller(1)
	//fp := filepath.Join(filepath.Dir(filename), "./seed/News_Category_Dataset_v3.json")
	fp := filepath.Join(filepath.Dir(filename), "./seed/articles_small.json")

	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	bar := progressbar.Default(1500)
	//bar := progressbar.Default(209527)
	defer bar.Finish()
	dec := json.NewDecoder(f)
	for {
		var a Article
		if err := dec.Decode(&a); err != nil {
			log.Printf("error: failed to decode article: %v\n", err)
			break
		}
		if err := r.InsertArticle(context.Background(), &a); err != nil {
			log.Printf("error: failed to insert article: %v\n", err)
			return err
		}
		bar.Add64(1)
	}

	return nil
}

func migrate(db *sql.DB) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

type encoder func(string) string

func newEncoder() (encoder, error) {
	enc, err := tasks.LoadModelForTextEncoding(&tasks.Config{ModelsDir: cacheDir(), ModelName: embeddingModel})
	if err != nil {
		return nil, err
	}

	encodeFn := func(text string) string {
		result, err := enc.Encode(context.Background(), text, int(bert.MeanPooling))
		if err != nil {
			log.Printf("failed to encode text: %v\n", err)
			return ""
		}
		b, err := json.Marshal(result.Vector.Data().F64()[:embeddingLimit])
		if err != nil {
			log.Printf("failed to marshal embedding: %v\n", err)
			return ""
		}
		return string(b)
	}

	return encodeFn, nil
}

var _cacheDir string

func cacheDir() string {
	if _cacheDir != "" {
		return _cacheDir
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("failed to get user cache dir: %v\n", err)
	}

	_cacheDir := path.Join(cacheDir, "sqlite-vss-examples")
	if !dirExists(_cacheDir) {
		if err := os.MkdirAll(_cacheDir, 0755); err != nil {
			log.Fatalf("failed to create cache directory: %v\n", err)
		}
	}

	return _cacheDir
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		log.Printf("checking: dirname does not exist: %s\n", dirname)
		return false
	}
	return info.IsDir()
}

func trackTime(msg string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s: %v\n", msg, time.Since(start))
	}
}

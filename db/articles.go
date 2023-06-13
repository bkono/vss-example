package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
)

type Article struct {
	ID          int64  `json:"id" bun:"rowid,pk,autoincrement"`
	Headline    string `json:"headline" bun:"headline"`
	Description string `json:"description" bun:"description"`
	Link        string `json:"link" bun:"link"`
	Category    string `json:"category" bun:"category"`
	Authors     string `json:"authors" bun:"authors"`
	Date        string `json:"date" bun:"date"`
}

type ArticleVector struct {
	bun.BaseModel `bun:"table:vss_articles"`

	ID                   int64  `json:"id" bun:"rowid,pk,autoincrement"`
	HeadlineEmbedding    string `json:"headline_embedding" bun:"headline_embedding"`
	DescriptionEmbedding string `json:"description_embedding" bun:"description_embedding"`
}

func (r *Repo) InsertArticle(ctx context.Context, article *Article) error {
	return r.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(article).
			Exec(ctx)
		if err != nil {
			return err
		}

		_, err = tx.NewInsert().
			Model(&ArticleVector{ID: article.ID}).
			Value("headline_embedding", "encode_embedding(?)", article.Headline).
			Value("description_embedding", "encode_embedding(?)", article.Description).
			Exec(ctx)

		return err
	})
}

func (r *Repo) CountArticles(ctx context.Context) (int, error) {
	return r.db.NewSelect().
		Model(&Article{}).
		Count(ctx)
}

func (r *Repo) ListArticles(ctx context.Context) ([]*Article, error) {
	var articles []*Article
	err := r.db.NewSelect().
		Model(&articles).
		Scan(ctx)
	return articles, err
}

func (r *Repo) ListArticleVectors(ctx context.Context) ([]*ArticleVector, error) {
	var vectors []*ArticleVector
	err := r.db.NewSelect().
		Model(&vectors).
		Scan(ctx)
	return vectors, err
}

type ArticleSearchResult struct {
	ID          int64   `json:"id" bun:"rowid"`
	Headline    string  `json:"headline" bun:"headline"`
	Description string  `json:"description" bun:"description"`
	Link        string  `json:"link" bun:"link"`
	Category    string  `json:"category" bun:"category"`
	Authors     string  `json:"authors" bun:"authors"`
	Date        string  `json:"date" bun:"date"`
	Distance    float64 `bun:"distance"`
}

func (r *Repo) SearchHeadlines(ctx context.Context, headline string, count int) []*ArticleSearchResult {
	var results []*ArticleSearchResult
	err := r.db.NewSelect().
		Model(&Article{}).
		ColumnExpr("article.*").
		ColumnExpr("article.rowid as rowid").
		ColumnExpr("vs.distance AS distance").
		Where("vss_search( vs.headline_embedding, vss_search_params(encode_embedding(?), ?))", headline, count).
		Join("JOIN vss_articles AS vs ON article.rowid = vs.rowid").
		OrderExpr("vs.distance ASC").
		Scan(ctx, &results)

	if err != nil {
		log.Printf("error: failed to search articles: %v", err)
		return results
	}

	return results
}

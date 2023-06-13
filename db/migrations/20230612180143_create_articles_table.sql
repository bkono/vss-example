-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS articles(
     headline text,
     description text,
     link text,
     category text,
     authors text,
     date
);

CREATE VIRTUAL TABLE vss_articles USING vss0(
    headline_embedding(384),
    description_embedding(384),
);

CREATE TRIGGER delete_vss_articles AFTER DELETE ON articles
BEGIN
    DELETE FROM vss_articles WHERE rowid = old.rowid;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vss_articles;
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd

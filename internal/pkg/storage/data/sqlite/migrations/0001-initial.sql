-- +migrate Up
CREATE TABLE files (
    id TEXT,
    filename TEXT,
    document_date DATETIME,
    hash TEXT
);
CREATE UNIQUE INDEX ix_files_id ON files (id);

CREATE TABLE file_metadata (
    id TEXT,
    hash TEXT,
    file_size INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX ix_file_metadata_id ON file_metadata(id);
CREATE UNIQUE INDEX ix_file_metadata_hash ON file_metadata(hash);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    system INTEGER
);
CREATE UNIQUE INDEX ix_tags_name ON tags(name);
INSERT INTO tags (name, system) VALUES ('unfiled', 1);

CREATE TABLE file_tags (
    file_id TEXT,
    tag_id INTEGER,
    FOREIGN KEY(file_id) REFERENCES files(id),
    FOREIGN KEY(tag_id) REFERENCES tags(id)
);
CREATE INDEX ix_file_tags_file_id ON file_tags (file_id);
CREATE INDEX ix_file_tags_tag_id ON file_tags (tag_id);

-- +migrate Down
DROP TABLE file_tags;
DROP TABLE tags;
DROP TABLE files;
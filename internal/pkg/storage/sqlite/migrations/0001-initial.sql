-- +migrate Up
CREATE TABLE files (
       id TEXT,
       hash TEXT,
       filename TEXT,
       document_date DATETIME,
       file_size INTEGER
);
CREATE UNIQUE INDEX ix_files_id ON files (id);
CREATE UNIQUE INDEX ix_files_hash ON files(hash);

CREATE TABLE tags (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name TEXT,
       system INTEGER
);
CREATE UNIQUE INDEX ix_tags_name ON tags(name);
INSERT INTO tags (name, system) VALUES ('unfiled', 1);

CREATE TABLE file_tags (
       file_id INTEGER,
       tag_id INTEGER
);
CREATE INDEX ix_file_tags_file_id ON file_tags (file_id);
CREATE INDEX ix_file_tags_tag_id ON file_tags (tag_id);

-- +migrate Down
DROP TABLE files;

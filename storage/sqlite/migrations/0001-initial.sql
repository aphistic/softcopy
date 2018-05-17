-- +migrate Up
CREATE TABLE files (
       id text,
       hash text,
       filename text,
       document_date datetime
);
CREATE UNIQUE INDEX ix_files_id ON files (id);
CREATE UNIQUE INDEX ix_files_hash ON files(hash);

-- +migrate Down
DROP TABLE files;

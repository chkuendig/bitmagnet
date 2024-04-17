-- +goose Up
-- +goose StatementBegin
create table releases (
    title text primary key,
    nfo text,
    size text,
    files text,
    filename text,
    nuked boolean,
    nukereason text,
    category text,
    created timestamp with time zone not null,
    source text,
    requestid text,
    groupname text,
    nzedbpre_dump timestamp with time zone not null
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists releases cascade;

-- +goose StatementEnd
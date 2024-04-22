-- +goose Up
-- +goose StatementBegin
create table releases (
    title text not null,
    nfo text,
    size text,
    files text,
    filename text,
    nuked int,
    nukereason text,
    category text,
    created timestamp with time zone not null,
    source text,
    requestid text,
    groupname text,
    nzedbpre_dump timestamp with time zone not null,
    PRIMARY KEY (title,nfo,size,files,filename,nuked,nukereason,category,created,source,requestid,groupname)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists releases cascade;

-- +goose StatementEnd
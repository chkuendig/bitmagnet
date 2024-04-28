-- +goose Up
-- +goose StatementBegin
-- Table: public.releases

-- DROP TABLE IF EXISTS public.releases;

CREATE TABLE IF NOT EXISTS public.releases
(
    name text COLLATE pg_catalog."default" NOT NULL,
    created date,
    nuked integer,
    category text COLLATE pg_catalog."default",
    nfo text COLLATE pg_catalog."default",
    CONSTRAINT releases_pkey PRIMARY KEY (name)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.releases
    OWNER to postgres;
    
-- FUNCTION: public.update_create_release()

-- DROP FUNCTION IF EXISTS public.update_create_release();

CREATE OR REPLACE FUNCTION public.update_create_release()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
BEGIN
     INSERT INTO         releases(name, category, nuked, created, nfo)
        VALUES(new.name, new.category, new.nuked, new.created, new.nfo) 
	-- ON CONFLICT DO NOTHING;
 ON CONFLICT ON CONSTRAINT releases_pkey
DO UPDATE
SET
  created = CASE WHEN (releases.created > new.created) OR (releases.created IS NULL) THEN new.created ELSE releases.created END,
  category =  CASE WHEN (LENGTH(releases.category) < LENGTH(new.category)) or (releases.category IS NULL )THEN new.category ELSE releases.category END,
  nuked = CASE WHEN (releases.nuked < new.nuked )or (releases.nuked IS NULL  )THEN new.nuked ELSE releases.nuked END,
  nfo = CASE WHEN (LENGTH(releases.nfo) < LENGTH(new.nfo)) or (releases.nfo IS NULL )THEN new.nfo ELSE releases.nfo END;

           RETURN new;
END;
$BODY$;

ALTER FUNCTION public.update_create_release()
    OWNER TO postgres;


-- Table: public.release_pre

-- DROP TABLE IF EXISTS public.release_pre;

CREATE TABLE IF NOT EXISTS public.release_pre
(
    name text COLLATE pg_catalog."default" NOT NULL,
    nfo text COLLATE pg_catalog."default",
    size text COLLATE pg_catalog."default",
    files text COLLATE pg_catalog."default",
    filename text COLLATE pg_catalog."default",
    nuked integer,
    nukereason text COLLATE pg_catalog."default",
    category text COLLATE pg_catalog."default",
    created timestamp with time zone NOT NULL,
    source text COLLATE pg_catalog."default" NOT NULL,
    requestid text COLLATE pg_catalog."default",
    groupname text COLLATE pg_catalog."default",
    nzedbpre_dump timestamp with time zone,
    CONSTRAINT unique_pre UNIQUE NULLS NOT DISTINCT (name, nuked, source, nukereason),
    CONSTRAINT release_name FOREIGN KEY (name)
        REFERENCES public.releases (name) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.release_pre
    OWNER to postgres;
-- Index: fki_release_name

-- DROP INDEX IF EXISTS public.fki_release_name;

CREATE INDEX IF NOT EXISTS fki_release_name
    ON public.release_pre USING btree
    (name COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: name_created_nuked_category_nfo

-- DROP INDEX IF EXISTS public.name_created_nuked_category_nfo;

CREATE INDEX IF NOT EXISTS name_created_nuked_category_nfo
    ON public.release_pre USING btree
    (name COLLATE pg_catalog."default" ASC NULLS LAST)
    INCLUDE(name, created, nuked, category, nfo)
    WITH (deduplicate_items=True)
    TABLESPACE pg_default;
-- Index: nzedbpre_dump

-- DROP INDEX IF EXISTS public.nzedbpre_dump;

CREATE INDEX IF NOT EXISTS nzedbpre_dump
    ON public.release_pre USING btree
    (nzedbpre_dump DESC NULLS FIRST)
    INCLUDE(nzedbpre_dump)
    TABLESPACE pg_default;

-- Trigger: update_create_release

-- DROP TRIGGER IF EXISTS update_create_release ON public.release_pre;


-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS releases cascade;
DROP TABLE IF EXISTS release_pre cascade;
DROP FUNCTION IF EXISTS update_create_release();
-- +goose StatementEnd
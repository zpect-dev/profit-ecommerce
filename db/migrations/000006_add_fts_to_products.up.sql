ALTER TABLE art ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('spanish', coalesce(co_art, '')), 'A') ||
    setweight(to_tsvector('spanish', coalesce(art_des, '')), 'B') ||
    setweight(to_tsvector('spanish', coalesce(campo4, '')), 'C')
) STORED;

CREATE INDEX idx_art_search_vector ON art USING GIN(search_vector);

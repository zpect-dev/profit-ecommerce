DROP INDEX IF EXISTS idx_art_search_vector;
ALTER TABLE art DROP COLUMN IF EXISTS search_vector;

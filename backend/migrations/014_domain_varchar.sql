-- 014_domain_varchar.sql
-- Convert domain column from ENUM to VARCHAR to support dynamic domain catalog

BEGIN;

-- Step 1: Drop default (if any) and convert the column
ALTER TABLE experiences ALTER COLUMN domain DROP DEFAULT;
ALTER TABLE experiences ALTER COLUMN domain TYPE VARCHAR(100);

-- Step 2: Also convert domains table name column (was created as VARCHAR already, just verify)
-- No change needed — domains table was created with VARCHAR

-- Step 3: Update any old domain values
-- 'life' -> 'life-philosophy'
UPDATE experiences SET domain = 'life-philosophy' WHERE domain = 'life';
-- 'emotion' -> 'wellness'  
UPDATE experiences SET domain = 'wellness' WHERE domain = 'emotion';

-- Step 4: Drop the old ENUM type
DROP TYPE IF EXISTS domain_type;

COMMIT;

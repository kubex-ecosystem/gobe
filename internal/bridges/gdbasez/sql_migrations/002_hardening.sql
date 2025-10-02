-- 1) [BUG] DB alvo: você setou CONNECT ON DATABASE kubex_db,
--    mas pelo seu container é "gdbase" (POSTGRES_DB=gdbase).
--    Ajuste o nome do DATABASE aqui:
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_database WHERE datname = 'gdbase') THEN
    EXECUTE 'GRANT CONNECT ON DATABASE gdbase TO readonly, readwrite, admin';
  END IF;
END$$;

-- 2) Garantir ownership “coeso” e restringir o schema public (evita bagunça de CREATE por PUBLIC)
ALTER SCHEMA public OWNER TO admin;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO readonly, readwrite, admin;

-- 3) Grants em SEQUENCES (muita gente esquece; sem isso, INSERT com seq falha para readwrite):
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO readonly, readwrite, admin;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO readonly;
GRANT UPDATE ON ALL SEQUENCES IN SCHEMA public TO readwrite, admin;

-- 4) Grants em FUTUROS objetos (tabelas, sequências, funções, tipos) — idempotente por owner atual.
DO $$
DECLARE
  r RECORD;

BEGIN
  FOR r IN SELECT n.nspname AS nsp
           FROM pg_namespace n
           WHERE n.nspname IN ('public')
  LOOP
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT ON TABLES TO readonly', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO readwrite', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT ALL ON TABLES TO admin', r.nsp);

    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT USAGE ON SEQUENCES TO readonly', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT USAGE,UPDATE ON SEQUENCES TO readwrite', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT ALL ON SEQUENCES TO admin', r.nsp);

    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT EXECUTE ON FUNCTIONS TO readwrite, readonly', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT ALL ON FUNCTIONS TO admin', r.nsp);

    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT USAGE ON TYPES TO readonly, readwrite', r.nsp);
    EXECUTE format('ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT ALL ON TYPES TO admin', r.nsp);
  END LOOP;
END$$;

-- 5) Extensões: você ativou uuid-ossp + pgcrypto.
--    Em Postgres 13+ dá pra usar gen_random_uuid() de pgcrypto e aposentar uuid-ossp.
--    Mantendo sua escolha, só garanta ownership consistente:
ALTER EXTENSION "uuid-ossp" OWNER TO admin;
ALTER EXTENSION "pgcrypto"  OWNER TO admin;
ALTER EXTENSION "pg_trgm"   OWNER TO admin;
ALTER EXTENSION "btree_gist" OWNER TO admin;
ALTER EXTENSION "fuzzystrmatch" OWNER TO admin;
ALTER EXTENSION "hstore" OWNER TO admin;

-- 6) Índices/tsvector: você fez trigger (ok). Alternativa moderna: coluna gerada.
--    Se quiser migrar depois:
--    ALTER TABLE products ADD COLUMN search_vector_g tsvector GENERATED ALWAYS AS (
--      setweight(to_tsvector('portuguese', coalesce(name,'')), 'A') ||
--      setweight(to_tsvector('portuguese', coalesce(sku,'')), 'A') ||
--      setweight(to_tsvector('portuguese', coalesce(barcode,'')), 'A') ||
--      setweight(to_tsvector('portuguese', coalesce(description,'')), 'C') ||
--      setweight(to_tsvector('portuguese', coalesce(brand,'')), 'B') ||
--      setweight(to_tsvector('portuguese', coalesce(manufacturer,'')), 'B')
--    ) STORED;
--    CREATE INDEX IF NOT EXISTS idx_products_search_gin ON products USING GIN (search_vector_g);

-- 7) Case-insensitive em username/email (qualidade de vida):
CREATE EXTENSION IF NOT EXISTS citext;
ALTER TABLE users ALTER COLUMN email TYPE citext;
ALTER TABLE users ALTER COLUMN username TYPE citext;

-- 8) Usuários default com senha plana: em dev tudo bem;
--    em prod, prefira SCRAM com políticas. Como mitigação:
ALTER ROLE user_readonly PASSWORD NULL;   -- força setar depois (se quiser)
ALTER ROLE user_readwrite PASSWORD NULL;  -- idem
-- Ou mantenha e rotacione via app/secret manager.

-- 9) (Opcional) RLS toggles para quando acoplar Supabase/Auth:
-- ALTER TABLE users ENABLE ROW LEVEL SECURITY;
-- CREATE POLICY users_self ON users
--   FOR SELECT USING (id::text = current_setting('request.jwt.claims', true)::json->>'sub');

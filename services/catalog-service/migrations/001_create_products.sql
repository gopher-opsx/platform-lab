CREATE TABLE IF NOT EXISTS products (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL,
    price_cents  BIGINT NOT NULL CHECK (price_cents >= 0),
    currency     CHAR(3) NOT NULL,
    image_url    TEXT NOT NULL,
    in_stock     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE Table if not exists "urls" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    long_url TEXT NOT NULL UNIQUE,
    short_code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
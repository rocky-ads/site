-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    seed_ad_file TEXT
);

-- Chains table (represents field chains within categories)
CREATE TABLE IF NOT EXISTS chains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    spec_table TEXT, -- NULL if chain has no spec fields
    chain_file TEXT, -- NULL if chain has no spec fields
    chain_index INTEGER NOT NULL, -- Order of chain within category
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, chain_index)
);

-- Fields table
CREATE TABLE IF NOT EXISTS fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL
);

-- Chain fields table (defines fields for each chain and their relationships)
CREATE TABLE IF NOT EXISTS chain_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chain_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    next_in_chain INTEGER NOT NULL DEFAULT 0, -- Points to field_order of next field, 0 if last
    is_required INTEGER NOT NULL DEFAULT 0,
    field_order INTEGER NOT NULL, -- Order of field within chain
    FOREIGN KEY (chain_id) REFERENCES chains(id),
    FOREIGN KEY (field_id) REFERENCES fields(id),
    UNIQUE(chain_id, field_id)
);

-- Ads table
-- Uses JSON for location to leverage SQLite JSON functions
CREATE TABLE IF NOT EXISTS ads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL,
    created_at TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    image_count INTEGER DEFAULT 0,
    location TEXT, -- JSON column for location data
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- Ad values for spec fields (supports both single and multi-value fields, one row per value)
CREATE TABLE IF NOT EXISTS ad_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ad_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    value TEXT NOT NULL CHECK (value != ''),
    FOREIGN KEY (ad_id) REFERENCES ads(id) ON DELETE CASCADE,
    FOREIGN KEY (field_id) REFERENCES fields(id),
    UNIQUE(ad_id, field_id, value)
);

-- Spec combinations: make->model (for Bicycles)
CREATE TABLE IF NOT EXISTS spec_bicycle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    make TEXT NOT NULL CHECK (make != ''),
    model TEXT NOT NULL CHECK (model != ''),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, make, model)
);

-- Spec combinations: make->year->model (for Agricultural Equipment)
CREATE TABLE IF NOT EXISTS spec_ag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    make TEXT NOT NULL CHECK (make != ''),
    year TEXT NOT NULL CHECK (year != ''),
    model TEXT NOT NULL CHECK (model != ''),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, make, year, model)
);

-- Spec combinations: make->year->model->engine (for Cars & Trucks, Motorcycles)
CREATE TABLE IF NOT EXISTS spec_vehicle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    make TEXT NOT NULL CHECK (make != ''),
    year TEXT NOT NULL CHECK (year != ''),
    model TEXT NOT NULL CHECK (model != ''),
    engine TEXT NOT NULL CHECK (engine != ''),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, make, year, model, engine)
);

-- Part combinations (part_category->part_subcategory)
CREATE TABLE IF NOT EXISTS spec_part (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    part_category TEXT CHECK (part_category IS NULL OR part_category != ''),
    part_subcategory TEXT CHECK (part_subcategory IS NULL OR part_subcategory != ''),
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, part_category, part_subcategory)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_chains_category ON chains(category_id);
CREATE INDEX IF NOT EXISTS idx_chain_fields_chain ON chain_fields(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_fields_field ON chain_fields(field_id);
CREATE INDEX IF NOT EXISTS idx_ads_category ON ads(category_id);
CREATE INDEX IF NOT EXISTS idx_ad_values_ad ON ad_values(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_values_field ON ad_values(field_id);
-- Composite index for efficient lookups (critical for search performance)
CREATE INDEX IF NOT EXISTS idx_ad_values_lookup ON ad_values(ad_id, field_id, value);


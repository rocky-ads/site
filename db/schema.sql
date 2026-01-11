-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    is_admin INTEGER NOT NULL DEFAULT 0,

    -- Encrypted user name
    encrypted_name BYTEA NOT NULL,
    name_nonce BYTEA NOT NULL,
    name_hash TEXT NOT NULL UNIQUE,

    -- Encrypted password
    password_hash TEXT NOT NULL,
    password_salt TEXT NOT NULL,
    password_algo TEXT NOT NULL DEFAULT 'argon2id',

    -- Encrypted phone
    encrypted_phone BYTEA NOT NULL,
    phone_nonce BYTEA NOT NULL,
    phone_hash TEXT NOT NULL UNIQUE,
    phone_verified INTEGER NOT NULL DEFAULT 0,
    sms_opted_out INTEGER NOT NULL DEFAULT 0,

    -- Encrypted email
    encrypted_email BYTEA NOT NULL,
    email_nonce BYTEA NOT NULL,
    email_hash TEXT UNIQUE,

    -- Notification settings
    notification_method TEXT NOT NULL DEFAULT 'sms',
    summary_notifications_enabled INTEGER DEFAULT 0,
    summary_notifications_frequency TEXT DEFAULT 'daily',
    last_sms_sent_at TIMESTAMP
);
CREATE INDEX idx_user_deleted_at ON users(deleted_at);
CREATE INDEX idx_user_name_hash ON users(name_hash);
CREATE INDEX idx_user_phone_hash ON users(phone_hash);
CREATE INDEX idx_users_last_sms_sent ON users(last_sms_sent_at);

-- Phone verification codes table
CREATE TABLE phone_verification (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_e64 TEXT NOT NULL,
    verification_code TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_phone_verification_phone_e64 ON phone_verification(phone_e64);
CREATE INDEX idx_phone_verification_code ON phone_verification(verification_code);
CREATE INDEX idx_phone_verification_created_at ON phone_verification(created_at);

-- Locations table
CREATE TABLE locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    raw_text TEXT UNIQUE NOT NULL CHECK (length(raw_text) > 0),
    city TEXT NOT NULL,
    admin_area TEXT NOT NULL,
    country TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL
);
CREATE INDEX idx_location_raw_text ON locations(raw_text);

-- Categories table
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    seed_ad_file TEXT,
    image_file TEXT
);

-- Chains table (represents field chains within categories)
CREATE TABLE chains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    spec_table TEXT, -- NULL if chain has no spec fields
    chain_file TEXT, -- NULL if chain has no spec fields
    chain_index INTEGER NOT NULL, -- Order of chain within category
    FOREIGN KEY (category_id) REFERENCES categories(id),
    UNIQUE(category_id, chain_index)
);
CREATE INDEX idx_chains_category ON chains(category_id);

-- Fields table
CREATE TABLE fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL
);

-- Chain fields table (defines fields for each chain and their relationships)
CREATE TABLE chain_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chain_id INTEGER NOT NULL REFERENCES chains(id),
    field_id INTEGER NOT NULL REFERENCES fields(id),
    next_in_chain INTEGER NOT NULL DEFAULT 0, -- Points to field_order of next field, 0 if last
    is_required INTEGER NOT NULL DEFAULT 0,
    field_order INTEGER NOT NULL, -- Order of field within chain
    UNIQUE(chain_id, field_id)
);
CREATE INDEX idx_chain_fields_chain ON chain_fields(chain_id);
CREATE INDEX idx_chain_fields_field ON chain_fields(field_id);

-- Ads table
CREATE TABLE ads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    price INTEGER NOT NULL, -- in cents
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id),
    image_count INTEGER DEFAULT 0,
    location_id INTEGER REFERENCES locations(id)
);
CREATE INDEX idx_ads_category ON ads(category_id);

-- Bookmarks table
CREATE TABLE bookmarks (
    user_id INTEGER NOT NULL REFERENCES users(id),
    ad_id INTEGER NOT NULL REFERENCES ads(id),
    bookmarked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, ad_id)
);
CREATE INDEX idx_bookmarks_user ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_ad ON bookmarks(ad_id);

-- Ad values for spec fields (supports both single and multi-value fields, one row per value)
CREATE TABLE ad_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ad_id INTEGER NOT NULL REFERENCES ads(id),
    field_id INTEGER NOT NULL REFERENCES fields(id),
    value TEXT NOT NULL CHECK (value != ''),
    UNIQUE(ad_id, field_id, value)
);
CREATE INDEX idx_ad_values_ad ON ad_values(ad_id);
CREATE INDEX idx_ad_values_field ON ad_values(field_id);
CREATE INDEX idx_ad_values_lookup ON ad_values(ad_id, field_id, value);

-- Spec combinations: make->model (for Bicycles)
CREATE TABLE spec_bicycle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    make TEXT NOT NULL CHECK (make != ''),
    model TEXT NOT NULL CHECK (model != ''),
    UNIQUE(category_id, make, model)
);

-- Spec combinations: make->year->model (for Agricultural Equipment)
CREATE TABLE spec_ag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    make TEXT NOT NULL CHECK (make != ''),
    year TEXT NOT NULL CHECK (year != ''),
    model TEXT NOT NULL CHECK (model != ''),
    UNIQUE(category_id, make, year, model)
);

-- Spec combinations: make->year->model->engine (for Cars & Trucks, Motorcycles)
CREATE TABLE spec_vehicle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    make TEXT NOT NULL CHECK (make != ''),
    year TEXT NOT NULL CHECK (year != ''),
    model TEXT NOT NULL CHECK (model != ''),
    engine TEXT NOT NULL CHECK (engine != ''),
    UNIQUE(category_id, make, year, model, engine)
);

-- Part combinations (part_category->part_subcategory)
CREATE TABLE spec_part (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    part_category TEXT CHECK (part_category IS NULL OR part_category != ''),
    part_subcategory TEXT CHECK (part_subcategory IS NULL OR part_subcategory != ''),
    UNIQUE(category_id, part_category, part_subcategory)
);

-- Conversations table
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ad_id INTEGER NOT NULL REFERENCES ads(id),
    owner_id INTEGER NOT NULL REFERENCES users(id),
    enquirer_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    owner_has_unread INTEGER NOT NULL DEFAULT 0,
    enquirer_has_unread INTEGER NOT NULL DEFAULT 0,
    is_public INTEGER NOT NULL DEFAULT 0,
    UNIQUE(ad_id, enquirer_id)
);
CREATE INDEX idx_conversations_owner ON conversations(owner_id);
CREATE INDEX idx_conversations_enquirer ON conversations(enquirer_id);
CREATE INDEX idx_conversations_ad ON conversations(ad_id);
CREATE INDEX idx_conversations_updated_at ON conversations(updated_at);
CREATE INDEX idx_conversations_is_public ON conversations(is_public);

-- Messages table
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    sender_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL CHECK (content != ''),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- Rocks table
CREATE TABLE rocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    thrown_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, conversation_id)
);
CREATE INDEX idx_rocks_user ON rocks(user_id);
CREATE INDEX idx_rocks_conversation ON rocks(conversation_id);

-- SMS notification queue table
CREATE TABLE sms_notification_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipient_user_id INTEGER NOT NULL REFERENCES users(id),
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending',
    FOREIGN KEY (recipient_user_id) REFERENCES users(id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
CREATE INDEX idx_sms_queue_recipient_status ON sms_notification_queue(recipient_user_id, status);
CREATE INDEX idx_sms_queue_status_created ON sms_notification_queue(status, created_at);
CREATE INDEX idx_sms_queue_processed_at ON sms_notification_queue(processed_at);
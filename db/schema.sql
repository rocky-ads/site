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

-- Ads table
CREATE TABLE ads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    price INTEGER NOT NULL DEFAULT 0, -- whole units of price_currency; 0 means free
    price_currency TEXT NOT NULL DEFAULT 'USD',
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

-- Conversations table
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ad_id INTEGER NOT NULL REFERENCES ads(id),
    owner_id INTEGER NOT NULL REFERENCES users(id),
    enquirer_id INTEGER NOT NULL REFERENCES users(id),
    owner_has_unread INTEGER NOT NULL DEFAULT 0,
    enquirer_has_unread INTEGER NOT NULL DEFAULT 0,
    egg_thrower_id INTEGER REFERENCES users(id), -- NULL = no egg (private), NOT NULL = public, owner_id = bound to enquirer, enquirer_id = bound to ad
    egg_thrown_at TIMESTAMP, -- Only valid if egg_thrower_id IS NOT NULL
    UNIQUE(ad_id, enquirer_id)
);
CREATE INDEX idx_conversations_owner ON conversations(owner_id);
CREATE INDEX idx_conversations_enquirer ON conversations(enquirer_id);
CREATE INDEX idx_conversations_ad ON conversations(ad_id);
CREATE INDEX idx_conversations_egg_thrower ON conversations(egg_thrower_id);

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
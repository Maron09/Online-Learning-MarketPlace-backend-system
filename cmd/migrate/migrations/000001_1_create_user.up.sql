-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     first_name VARCHAR(255) NOT NULL,
--     last_name VARCHAR(255) NOT NULL,
--     email VARCHAR(255) NOT NULL UNIQUE,
--     password VARCHAR(255) NOT NULL,
--     role SMALLINT DEFAULT 2,
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     last_login TIMESTAMPTZ,
--     is_active BOOLEAN DEFAULT false,
--     otp VARCHAR(6),
--     otp_expires_at TIMESTAMPTZ
-- );
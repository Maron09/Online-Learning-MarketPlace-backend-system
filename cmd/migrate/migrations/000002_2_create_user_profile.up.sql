-- CREATE TABLE user_profiles (
--     id SERIAL PRIMARY KEY,
--     user_id INT REFERENCES users(id) ON DELETE CASCADE,
--     profile_picture VARCHAR(255),
--     country VARCHAR(100),
--     created_at TIMESTAMPTZ DEFAULT NOW(),
--     modified_at TIMESTAMPTZ DEFAULT NOW()
-- );
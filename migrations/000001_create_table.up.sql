-- Agar kerak bo'lsa, role_user turini o'chirib tashlash
DROP TYPE IF EXISTS role_user;

-- Role_user turini yaratish
CREATE TYPE role_user AS ENUM ('admin', 'user');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,     
    phone_number VARCHAR(15) NOT NULL,      
    password VARCHAR(255) NOT NULL,          
    full_name VARCHAR(100),                       
    profile_picture VARCHAR(255),                
    bio TEXT,                                     
    role role_user DEFAULT 'user',                
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at BIGINT DEFAULT 0
);


CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$   
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

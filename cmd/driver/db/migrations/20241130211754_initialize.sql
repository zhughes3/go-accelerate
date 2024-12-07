-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION generate_random_id()
RETURNS TEXT AS $$
DECLARE
chars TEXT := 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_';
    result TEXT := '';
BEGIN
FOR i IN 1..12 LOOP
        result := result || substr(chars, floor(random() * 62 + 1)::int, 1);
END LOOP;
RETURN result;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION generate_primary_key_trigger()
RETURNS TRIGGER AS $$
DECLARE
table_name TEXT := TG_TABLE_NAME; -- Get the name of the table that triggered this function
    key_exists BOOLEAN;
BEGIN
    LOOP
        -- Generate a random ID
NEW.id := generate_random_id();

        -- Check for uniqueness using dynamic SQL
EXECUTE format(
        'SELECT EXISTS (SELECT 1 FROM %I WHERE id = $1)',
        table_name
        )
    INTO key_exists
    USING NEW.id;

-- Exit the loop if the ID is unique
EXIT WHEN NOT key_exists;
END LOOP;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE users(
    id TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    PRIMARY KEY(id)
);

CREATE TRIGGER set_random_id
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION generate_primary_key_trigger();

CREATE TABLE timelines(
    id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TRIGGER set_random_id
    BEFORE INSERT ON timelines
    FOR EACH ROW
    EXECUTE FUNCTION generate_primary_key_trigger();

CREATE TABLE events(
    id TEXT NOT NULL,
    timeline_id TEXT NOT NULL,
    title TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    description TEXT,
    content TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    PRIMARY KEY (id),
    FOREIGN KEY (timeline_id) REFERENCES timelines (id) ON DELETE CASCADE ON UPDATE CASCADE,
    -- Add a unique index on (event_id, timeline_id) since we will commonly query this table using both fields.
    UNIQUE (id, timeline_id)
);
CREATE TRIGGER set_random_id
    BEFORE INSERT ON events
    FOR EACH ROW
    EXECUTE FUNCTION generate_primary_key_trigger();

-- +goose Down

DROP TABLE events;
DROP TABLE timelines;
DROP TABLE users;
DROP FUNCTION generate_primary_key_trigger();
DROP FUNCTION generate_random_id();




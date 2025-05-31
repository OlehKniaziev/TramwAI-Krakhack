CREATE TABLE IF NOT EXISTS Events (
        id SERIAL PRIMARY KEY,
        date date NOT NULL,
        title text NOT NULL,
        description text NOT NULL,
        link text NOT NULL,
        location text NOT NULL
);

CREATE TABLE IF NOT EXISTS Keywords (
        id SERIAL PRIMARY KEY,
        keyword text NOT NULL,
        event_id int NOT NULL REFERENCES Events(id)
);

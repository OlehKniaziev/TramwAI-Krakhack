CREATE TABLE IF NOT EXISTS Events (
        id SERIAL PRIMARY KEY,
        date text NOT NULL,
        title text NOT NULL,
        description text NOT NULL,
        link text NOT NULL
);

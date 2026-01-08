CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     github_login TEXT UNIQUE NOT NULL,
                                     telegram_chat_id BIGINT,
                                     created_at TIMESTAMP DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS repositories (
                                            id SERIAL PRIMARY KEY,
                                            full_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS pull_requests (
                                             id SERIAL PRIMARY KEY,
                                             github_pr_id BIGINT NOT NULL,
                                             repository_id INT REFERENCES repositories(id),
    author_login TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
    );

package storage

import (
	"context"
	"log"
	"time"

	"github.com/fncg/ReviewerService/internal/github"
	"github.com/fncg/ReviewerService/internal/telegram"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(dsn string) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) ensureRepository(fullName string) (int, error) {
	ctx := context.Background()
	var repoID int
	err := p.pool.QueryRow(ctx,
		`INSERT INTO repositories (full_name) VALUES ($1)
         ON CONFLICT (full_name) DO UPDATE SET full_name=EXCLUDED.full_name
         RETURNING id`, fullName).Scan(&repoID)
	if err != nil {
		return 0, err
	}
	return repoID, nil
}

func (p *Postgres) SavePullRequest(event github.PullRequestEvent) (int, error) {
	ctx := context.Background()
	repoID, err := p.ensureRepository(event.Repository.FullName)
	if err != nil {
		return 0, err
	}

	var prID int
	err = p.pool.QueryRow(ctx,
		`INSERT INTO pull_requests (github_pr_id, repository_id, author_login)
         VALUES ($1, $2, $3)
         RETURNING id`,
		event.PullRequest.ID, repoID, event.PullRequest.User.Login).Scan(&prID)
	if err != nil {
		return 0, err
	}

	return prID, nil
}

func (p *Postgres) SelectReviewer(author string) (string, error) {
	ctx := context.Background()
	var reviewer string
	err := p.pool.QueryRow(ctx,
		`SELECT github_login
         FROM users
         WHERE github_login != $1
         ORDER BY RANDOM()
         LIMIT 1`, author).Scan(&reviewer)
	if err != nil {
		return "", err
	}
	return reviewer, nil
}

func (p *Postgres) AssignReviewer(prID int, reviewer string) error {
	ctx := context.Background()
	_, err := p.pool.Exec(ctx,
		`INSERT INTO review_assignments (pr_id, reviewer_login) VALUES ($1, $2)`,
		prID, reviewer)
	return err
}

func (p *Postgres) NotifyReviewer(reviewer string, message string, bot *telegram.Bot) {
	var chatID int64
	err := p.pool.QueryRow(context.Background(),
		`SELECT telegram_chat_id FROM users WHERE github_login=$1`, reviewer).Scan(&chatID)
	if err != nil {
		log.Println("Failed to get Telegram chat ID:", err)
		return
	}

	if chatID != 0 {
		bot.Notify(chatID, message)
	}
}

func (p *Postgres) HandlePR(event github.PullRequestEvent, bot *telegram.Bot) {
	prID, err := p.SavePullRequest(event)
	if err != nil {
		log.Println("Error saving PR:", err)
		return
	}

	reviewer, err := p.SelectReviewer(event.PullRequest.User.Login)
	if err != nil {
		log.Println("Error selecting reviewer:", err)
		return
	}

	if err := p.AssignReviewer(prID, reviewer); err != nil {
		log.Println("Error assigning reviewer:", err)
		return
	}

	log.Printf("PR %d saved, reviewer assigned: %s\n", prID, reviewer)

	msg := "You have been assigned to review PR: " + event.PullRequest.HTML
	p.NotifyReviewer(reviewer, msg, bot)
}

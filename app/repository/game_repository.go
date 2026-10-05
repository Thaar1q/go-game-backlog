package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"game_backlog_api/app/model"
)

// 1. Interface & Structs
type GameRepository interface {
	FindAll(ctx context.Context, q model.GameListQuery) ([]model.Game, int, error)
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Game, error)
	FindByID(ctx context.Context, id int) (model.Game, error)
	Create(ctx context.Context, g model.Game) (model.Game, error)
	Update(ctx context.Context, g model.Game) (model.Game, error)
	Delete(ctx context.Context, id int) error
}

// Whitelist
var gameSortColumns = map[string]string{
	"id":           "id",
	"title":        "title",
	"release_year": "release_year",
	"created_at":   "created_at",
}

const gameColumns = "id, title, platform, genre, release_year, created_at, updated_at"

type gamePostgresRepository struct {
	pool *pgxpool.Pool
}

// 2. Constructor
func NewGameRepository(pool *pgxpool.Pool) GameRepository {
	return &gamePostgresRepository{pool: pool}
}

// 3. Helper
func scanGame(row pgx.Row) (model.Game, error) {
	var g model.Game
	err := row.Scan(&g.ID, &g.Title, &g.Platform, &g.Genre, &g.ReleaseYear, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}

func buildGameFilter(q model.GameListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.Genre != "" {
		where += fmt.Sprintf(" AND genre ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Genre+"%")
	}
	if q.Platform != "" {
		where += fmt.Sprintf(" AND platform ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Platform+"%")
	}

	return where, args
}

// 4. Find All (List)
func (r *gamePostgresRepository) FindAll(ctx context.Context, q model.GameListQuery) ([]model.Game, int, error) {
	where, args := buildGameFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM games"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count games: %w", err)
	}

	sortCol := "id"
	if col, ok := gameSortColumns[q.Sort]; ok {
		sortCol = col
	}

	orderDir := "ASC"
	if q.Order == "desc" {
		orderDir = "DESC"
	}

	if q.Limit <= 0 {
		q.Limit = 10
	}
	offset := q.Offset()

	sqlText := fmt.Sprintf(
		`SELECT %s FROM games%s
         ORDER BY %s %s
         LIMIT $%d OFFSET $%d`,
		gameColumns, where, sortCol, orderDir, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, offset)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch games list: %w", err)
	}
	defer rows.Close()

	games := []model.Game{}
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan game row: %w", err)
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read games result: %w", err)
	}

	return games, total, nil
}

// 5. Find after Cursor
func (r *gamePostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Game, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND title ILIKE $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, limit+1)

	query := fmt.Sprintf(
		`SELECT %s FROM games%s
         ORDER BY created_at DESC, id DESC
         LIMIT $%d`,
		gameColumns, where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetch cursor games: %w", err)
	}
	defer rows.Close()

	games := []model.Game{}
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cursor game: %w", err)
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read cursor games: %w", err)
	}

	return games, nil
}

// 6. Find By ID (Get One)
func (r *gamePostgresRepository) FindByID(ctx context.Context, id int) (model.Game, error) {
	query := fmt.Sprintf(`SELECT %s FROM games WHERE id = $1`, gameColumns)
	g, err := scanGame(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Game{}, ErrNotFound
		}
		return model.Game{}, fmt.Errorf("fetch game: %w", err)
	}
	return g, nil
}

// 7. Create (Insert)
func (r *gamePostgresRepository) Create(ctx context.Context, g model.Game) (model.Game, error) {
	query := `INSERT INTO games (title, platform, genre, release_year)
              VALUES ($1, $2, $3, $4)
              RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, g.Title, g.Platform, g.Genre, g.ReleaseYear).
		Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Game{}, ErrDuplicate
		}
		return model.Game{}, fmt.Errorf("save game: %w", err)
	}
	return g, nil
}

// 8. Update (Replace/Patch)
func (r *gamePostgresRepository) Update(ctx context.Context, g model.Game) (model.Game, error) {
	query := fmt.Sprintf(
		`UPDATE games
         SET title = $1, platform = $2, genre = $3, release_year = $4, updated_at = NOW()
         WHERE id = $5
         RETURNING %s`,
		gameColumns,
	)

	updated, err := scanGame(r.pool.QueryRow(ctx, query, g.Title, g.Platform, g.Genre, g.ReleaseYear, g.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Game{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Game{}, ErrDuplicate
		}
		return model.Game{}, fmt.Errorf("update game: %w", err)
	}
	return updated, nil
}

// 9. Delete
func (r *gamePostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM games WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

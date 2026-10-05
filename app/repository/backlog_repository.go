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
type BacklogRepository interface {
	FindAllByUser(ctx context.Context, userID int, q model.BacklogListQuery) ([]model.BacklogResponse, int, error)
	FindByID(ctx context.Context, id int) (model.BacklogResponse, error)
	Create(ctx context.Context, b model.Backlog) (model.Backlog, error)
	Update(ctx context.Context, b model.Backlog) (model.Backlog, error)
	Delete(ctx context.Context, id int) error
}

// Whitelist
var backlogSortColumns = map[string]string{
	"id":           "b.id",
	"rating":       "b.rating",
	"hours_played": "b.hours_played",
	"updated_at":   "b.updated_at",
	"created_at":   "b.created_at",
}

const backlogJoinedColumns = `b.id, b.user_id, b.game_id, b.status, b.hours_played, b.rating, b.notes, b.created_at, b.updated_at,
       g.title, g.platform, g.genre`

type backlogPostgresRepository struct {
	pool *pgxpool.Pool
}

// 2. Constructor
func NewBacklogRepository(pool *pgxpool.Pool) BacklogRepository {
	return &backlogPostgresRepository{pool: pool}
}

// 3. Helper
func scanBacklogResponse(row pgx.Row) (model.BacklogResponse, error) {
	var res model.BacklogResponse
	err := row.Scan(
		&res.ID,
		&res.UserID,
		&res.GameID,
		&res.Status,
		&res.HoursPlayed,
		&res.Rating,
		&res.Notes,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.GameTitle,
		&res.GamePlatform,
		&res.GameGenre,
	)
	return res, err
}

// 4. Find All by User (List)
func (r *backlogPostgresRepository) FindAllByUser(ctx context.Context, userID int, q model.BacklogListQuery) ([]model.BacklogResponse, int, error) {
	where := " WHERE b.user_id = $1"
	args := []any{userID}

	if q.Status != "" {
		args = append(args, q.Status)
		where += fmt.Sprintf(" AND b.status = $%d", len(args))
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM backlogs b %s", where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count backlogs: %w", err)
	}

	sortCol := "b.id"
	if col, ok := backlogSortColumns[q.Sort]; ok {
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
		`SELECT %s
         FROM backlogs b
         INNER JOIN games g ON g.id = b.game_id
         %s
         ORDER BY %s %s
         LIMIT $%d OFFSET $%d`,
		backlogJoinedColumns, where, sortCol, orderDir, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, offset)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch backlog list: %w", err)
	}
	defer rows.Close()

	backlogs := []model.BacklogResponse{}
	for rows.Next() {
		b, err := scanBacklogResponse(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan backlog row: %w", err)
		}
		backlogs = append(backlogs, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read backlog result: %w", err)
	}

	return backlogs, total, nil
}

// 5. Find By ID (Get One)
func (r *backlogPostgresRepository) FindByID(ctx context.Context, id int) (model.BacklogResponse, error) {
	query := fmt.Sprintf(
		`SELECT %s
         FROM backlogs b
         INNER JOIN games g ON g.id = b.game_id
         WHERE b.id = $1`,
		backlogJoinedColumns,
	)

	res, err := scanBacklogResponse(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BacklogResponse{}, ErrNotFound
		}
		return model.BacklogResponse{}, fmt.Errorf("fetch backlog by id: %w", err)
	}
	return res, nil
}

// 6. Create (Insert)
func (r *backlogPostgresRepository) Create(ctx context.Context, b model.Backlog) (model.Backlog, error) {
	if b.Status == "" {
		b.Status = "backlog"
	}

	query := `INSERT INTO backlogs (user_id, game_id, status, hours_played, rating, notes)
              VALUES ($1, $2, $3, $4, $5, $6)
              RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, b.UserID, b.GameID, b.Status, b.HoursPlayed, b.Rating, b.Notes).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Backlog{}, ErrDuplicate
		}
		return model.Backlog{}, fmt.Errorf("save backlog: %w", err)
	}
	return b, nil
}

// 7. Update (Replace/Patch)
func (r *backlogPostgresRepository) Update(ctx context.Context, b model.Backlog) (model.Backlog, error) {
	query := `UPDATE backlogs
              SET status = $1, hours_played = $2, rating = $3, notes = $4, updated_at = NOW()
              WHERE id = $5
              RETURNING id, user_id, game_id, status, hours_played, rating, notes, created_at, updated_at`

	var updated model.Backlog
	err := r.pool.QueryRow(ctx, query, b.Status, b.HoursPlayed, b.Rating, b.Notes, b.ID).
		Scan(
			&updated.ID,
			&updated.UserID,
			&updated.GameID,
			&updated.Status,
			&updated.HoursPlayed,
			&updated.Rating,
			&updated.Notes,
			&updated.CreatedAt,
			&updated.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Backlog{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Backlog{}, ErrDuplicate
		}
		return model.Backlog{}, fmt.Errorf("update backlog: %w", err)
	}
	return updated, nil
}

// 8. Delete
func (r *backlogPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM backlogs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete backlog: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

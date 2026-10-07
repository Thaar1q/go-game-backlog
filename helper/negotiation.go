package helper

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"game_backlog_api/app/model"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))

	if accept == "" {
		return offered[0], nil
	}

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable(
			"requested format not available, choose from: " +
				strings.Join(offered, ", "))
	}

	return chosen, nil
}

func WriteGamesCSV(c *fiber.Ctx, games []model.Game) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="games.csv"`)

	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "title", "platform", "genre", "release_year", "created_at"}
	if err := writer.Write(header); err != nil {
		return Internal(err)
	}

	for _, g := range games {
		row := []string{
			strconv.Itoa(g.ID),
			g.Title,
			g.Platform,
			g.Genre,
			strconv.Itoa(g.ReleaseYear),
			g.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return Internal(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	return c.SendString(buffer.String())
}

func WriteBacklogCSV(c *fiber.Ctx, items []model.BacklogResponse) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="backlog.csv"`)

	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "game_id", "title", "platform", "genre", "status", "hours_played", "rating", "created_at"}
	if err := writer.Write(header); err != nil {
		return Internal(err)
	}

	for _, b := range items {
		ratingStr := ""
		if b.Rating != nil {
			ratingStr = strconv.Itoa(*b.Rating)
		}
		row := []string{
			strconv.Itoa(b.ID),
			strconv.Itoa(b.GameID),
			b.GameTitle,
			b.GamePlatform,
			b.GameGenre,
			b.Status,
			strconv.Itoa(b.HoursPlayed),
			ratingStr,
			b.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return Internal(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	return c.SendString(buffer.String())
}

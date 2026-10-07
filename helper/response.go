package helper

import (
	"game_backlog_api/app/model"

	"github.com/gofiber/fiber/v2"
)

func Success(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

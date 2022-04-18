package main

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func web() {
	app := fiber.New()
	//https://github.com/golang/go/issues/32350
	app.Use(func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}

		if strings.HasSuffix(c.OriginalURL(), ".js") {
			c.Response().Header.Set("Content-Type", "application/javascript")
		}

		return nil
	})
	app.Static("/", "./web/dist")

	api := app.Group("/api/")
	app.Use(cors.New())
	{
		api.Post("streamer", func(c *fiber.Ctx) error {
			p := new(Streamconfig)
			if err := c.BodyParser(p); err != nil {
				return err
			}
			streamerpool.Createstreamer(p)
			return c.SendString("ok")
		})
		api.Get("streamer", func(c *fiber.Ctx) error {
			return c.JSON(streamerpool)
		})
		api.Get("streamer/:name", func(c *fiber.Ctx) error {
			return c.JSON(streamerpool[c.Params("name")])
		})
		api.Delete("streamer/:name", func(c *fiber.Ctx) error {
			err := streamerpool.DeleteStreamer(c.Params("name"))
			if err != nil {
				return err
			}
			return c.SendString("ok")
		})
		api.Post("streamer/:name/start", func(c *fiber.Ctx) error {
			s, e := streamerpool[c.Params("name")]
			if !e {
				return errors.New("name no exited")
			}
			err := s.Startstream()
			if err != nil {
				return err
			}
			return c.SendString("ok")
		})
		api.Post("streamer/:name/stop", func(c *fiber.Ctx) error {
			s, e := streamerpool[c.Params("name")]
			if !e {
				return errors.New("name no exited")
			}
			err := s.Stopstream()
			if err != nil {
				return err
			}
			return c.SendString("ok")
		})
	}

	app.Listen(":13232")
}

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"planning-poker/internal/server/models"

	"github.com/gofiber/fiber/v2"
	"github.com/shareed2k/goth_fiber"
)

func (h *Handlers) AuthHandler(c *fiber.Ctx) error {
	err := goth_fiber.BeginAuthHandler(c)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (h *Handlers) AuthCallbackHandler(c *fiber.Ctx) error {
	user, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString(fmt.Sprintf("Authentication failed: %v", err))
	}

	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}
	sess.Set("user", models.User{
		Name:  user.Name,
		Email: user.Email,
	})
	save_err := sess.Save()
	if save_err != nil {
		log.Println(save_err)
	}

	return c.Redirect("/dashboard")
}

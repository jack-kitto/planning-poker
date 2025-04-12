package handlers

import (
	"fmt"
	"log"
	"net/http"

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
	sess.Set("user", map[string]string{
		"Name":     user.Name,
		"Email":    user.Email,
		"Provider": user.Provider,
	})
	save_err := sess.Save()
	if save_err != nil {
		log.Println(save_err)
	}

	return c.Redirect("/dashboard")
}

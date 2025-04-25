package handlers

import (
	"fmt"
	"net/http"
	"planning-poker/cmd/web/design/organisms"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (h *Handlers) StoryCreatePopupHandler(c *fiber.Ctx) error {
	sessionId := c.Params("id")
	return adaptor.HTTPHandler(templ.Handler(organisms.CreateStoryPopup(sessionId)))(c)
}

func (h *Handlers) CreateStoryHandler(c *fiber.Ctx) error {
	title := c.FormValue("title")
	sessionId := c.FormValue("sessionId")
	description := c.FormValue("description")
	index := ""
	session, err := h.DB.GetSession(sessionId)
	if err != nil {
		return err
	}
	_, err = h.DB.CreateStory(session, title, &description, index)
	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/test/session/%s", sessionId))
	return c.SendString("")
}

func (h *Handlers) UpdateStoryHandler(c *fiber.Ctx) error {
	session_id := c.FormValue("session_id")
	story_id := c.FormValue("story_id")
	title := c.FormValue("title")
	description := c.FormValue("description")
	err := h.DB.UpdateStory(story_id, title, description)
	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/test/session/%s", session_id))
	return c.SendString("")
}

func (h *Handlers) StoryEditPopupHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}
	story, err := h.DB.GetUserStory(id)
	if err != nil {
		return err
	}

	return adaptor.HTTPHandler(templ.Handler(organisms.UserStoryPopup(*story)))(c)
}

func (h *Handlers) GetUserStoryHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}
	story, err := h.DB.GetUserStory(id)
	if err != nil {
		return err
	}

	res := struct {
		UserStory any `json:"user_story"`
	}{
		UserStory: story,
	}

	return c.JSON(res)
}

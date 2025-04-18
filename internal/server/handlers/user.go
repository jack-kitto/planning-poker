package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) GetUserHandler(c *fiber.Ctx) error {
	params := c.Queries()
	user, err := h.DB.GetUser(params["email"])
	if err != nil {
		return err
	}
	return c.JSON(user)
}

func (h *Handlers) CreateUserHandler(c *fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")

	user, err := h.DB.CreateUser(name, email)
	if err != nil {
		return err
	}

	org, err := h.DB.CreateOrg(fmt.Sprintf("%s's Org", name), user)
	if err != nil {
		return err
	}

	_, err = h.DB.CreateOrgMember(org, user)
	if err != nil {
		return err
	}

	res := struct {
		User         any `json:"user"`
		Organisation any `json:"organisation"`
	}{
		User:         user,
		Organisation: org,
	}

	return c.JSON(res)
}

func (h *Handlers) UpdateUserHandler(c *fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	user, err := h.DB.UpdateUser(name, email)
	if err != nil {
		return err
	}
	return c.JSON(user)
}

package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"planning-poker/cmd/web/design/organisms"
	"planning-poker/cmd/web/pages"
	"planning-poker/internal/server/models"
	"sync"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/resend/resend-go/v2"
)

func (h *Handlers) TestSessionHandler(c *fiber.Ctx) error {
	session_name := c.FormValue("session-name")
	email := c.FormValue("email")
	// session,err := h.DB.CreateSession()
	user, err := h.DB.GetUserWithOrg(email)
	if err != nil {
		return err
	}
	// org, err := h.DB.GetOrg(user.OrganisationMembers[0].OrganisationID)
	// if err != nil {
	// 	return err
	// }
	session, err := h.DB.CreateSession(user.OrganisationMembers[0].Organisation, user, session_name)
	if err != nil {
		return err
	}

	sessionParticipant, err := h.DB.CreateSessionParticipant(session, user)
	if err != nil {
		return err
	}

	res := struct {
		User any `json:"user"`
		// Org  any `json:"org"`
		Session            any `json:"session"`
		SessionParticipant any `json:"session_participant"`
	}{
		User: user,
		// Org:  org,
		Session:            session,
		SessionParticipant: sessionParticipant,
	}

	return c.JSON(res)
}

func (h *Handlers) CreateSessionHandler(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}

	session_name := c.FormValue("session-name")

	userData := sess.Get("user")
	if userData == nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	user, ok := userData.(models.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}

	userWithOrg, err := h.DB.GetUserWithOrg(user.Email)
	if err != nil {
		return err
	}

	session, err := h.DB.CreateSession(userWithOrg.OrganisationMembers[0].Organisation, userWithOrg, session_name)
	if err != nil {
		return err
	}

	_, err = h.DB.CreateSessionParticipant(session, userWithOrg)
	if err != nil {
		return err
	}

	return c.Redirect(fmt.Sprintf("/session/%s", session.ID))
}

func (h *Handlers) GetSessionHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusInternalServerError).SendString("Could not find session identifier")
	}
	session, err := h.DB.GetSession(id)
	if err != nil {
		return err
	}
	res := struct {
		Session any `json:"session"`
	}{
		Session: session,
	}

	return c.JSON(res)
}

func (h *Handlers) EditSessionTitleFormHandler(c *fiber.Ctx) error {
	sessionId := c.Params("id")
	session, err := h.DB.GetSession(sessionId)
	if err != nil {
		return err
	}
	return adaptor.HTTPHandler(templ.Handler(organisms.EditSessionForm(session)))(c)
}

func (h *Handlers) EditSessionNameHandler(c *fiber.Ctx) error {
	name := c.FormValue("title")
	sessionId := c.Params("id")
	err := h.DB.UpdateSessionName(sessionId, name)
	if err != nil {
		return err
	}

	c.Set("HX-Redirect", fmt.Sprintf("/test/session/%s", sessionId))
	return c.SendString("")
}

func (h *Handlers) SessionTitleHandler(c *fiber.Ctx) error {
	sessionId := c.Params("id")
	session, err := h.DB.GetSession(sessionId)
	if err != nil {
		return err
	}
	return adaptor.HTTPHandler(templ.Handler(organisms.SessionTitle(session)))(c)
}

func (h *Handlers) InviteUserToSessionFormHandler(c *fiber.Ctx) error {
	sessionId := c.Params("id")
	return adaptor.HTTPHandler(templ.Handler(organisms.InviteUserPopup(sessionId)))(c)
}

func (h *Handlers) HandleSessionInvitation(c *fiber.Ctx) error {
	sessionId := c.FormValue("sessionId")
	email := c.FormValue("email")

	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	}

	session, err := h.DB.GetSession(sessionId)
	if err != nil {
		return err
	}

	userToAdd, err := h.DB.GetUser(email)
	if err != nil {
		return err
	}

	_, err = h.DB.CreateSessionParticipant(session, userToAdd)
	if err != nil {
		return err
	}

	userData := sess.Get("user")
	if userData == nil {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}

	user, ok := userData.(models.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).SendString("Invalid user data")
	}
	link := fmt.Sprintf("%s/session/%s", h.Config.BaseURL, sessionId)
	err = h.sendInvitation(email, fmt.Sprintf("%s has invited you to join %s", user.Name, session.Name), link, user.Name)
	if err != nil {
		return err
	}
	c.Set("HX-Redirect", fmt.Sprintf("/test/session/%s", sessionId))
	return c.SendString("")
}

func (h *Handlers) sendInvitation(to string, subject string, link string, from string) error {
	client := resend.NewClient(h.Config.EmailServerPassword)

	html := generateInviteEmailTemplate(from, link)

	params := &resend.SendEmailRequest{
		From:    h.Config.EmailFrom,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

func (h *Handlers) SessionPageHandler(c *fiber.Ctx) error {
	log.Printf("[SessionPageHandler] start")
	id := c.Params("id")
	if id == "" {
		log.Printf("[SessionPageHandler] no id found")
		return c.Status(http.StatusInternalServerError).SendString("Could not find session identifier")
	} else {
		log.Printf("[SessionPageHandler] found id %s", id)
	}

	sess, err := h.Store.Get(c)
	if err != nil {
		return err
	} else {
		log.Printf("[SessionPageHandler] found session")
	}
	userData := sess.Get("user")
	user, ok := userData.(models.User)
	if !ok {
		log.Printf("[SessionPageHandler] Could not find user")
		return c.Redirect("/login")
	}
	log.Printf("[SessionPageHandler] found user in session. Hi %s", user.Name)

	session, err := h.DB.GetSession(id)
	if err != nil {
		return err
	}

	return adaptor.HTTPHandler(templ.Handler(pages.SessionPage(session, &user)))(c)
}

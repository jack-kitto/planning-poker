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

var (
	sessionClients   = make(map[string]map[*websocket.Conn]bool)
	sessionClientsMu sync.Mutex
)

func (h *Handlers) SessionParticipantsWebsocketHandler(c *websocket.Conn) {
	log.Printf("[SessionParticipantsWebsocketHandler] start")
	user_id := c.Params("user_id")
	session_id := c.Params("session_id")
	log.Printf("Websocket params: user_id=%s, session_id=%s", user_id, session_id)

	// Add connection to sessionClients
	sessionClientsMu.Lock()
	if sessionClients[session_id] == nil {
		sessionClients[session_id] = make(map[*websocket.Conn]bool)
	}
	sessionClients[session_id][c] = true
	log.Printf("Session %s now has %d clients", session_id, len(sessionClients[session_id]))
	sessionClientsMu.Unlock()

	// Now call HandleSessionJoin
	h.HandleSessionJoin(session_id, user_id)

	defer func() {
		// Remove connection from sessionClients
		sessionClientsMu.Lock()
		delete(sessionClients[session_id], c)
		if len(sessionClients[session_id]) == 0 {
			delete(sessionClients, session_id)
		}
		log.Printf("Session %s now has %d clients", session_id, len(sessionClients[session_id]))
		sessionClientsMu.Unlock()

		h.HandleSessionLeave(session_id, user_id)
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Printf("Websocket read error: %v", err)
			break
		}
	}
}

func (h *Handlers) HandleSessionJoin(session_id string, user_id string) {
	log.Printf("[HandleSessionJoin] start - session_id=%s, user_id=%s", session_id, user_id)
	session, err := h.DB.GetSession(session_id)
	if err != nil {
		log.Printf("[HandleSessionJoin] GetSession error: %v", err)
		return
	}
	log.Printf("[HandleSessionJoin] Found session with ID: %s", session.ID)

	user, err := h.DB.GetUserById(user_id)
	if err != nil {
		log.Printf("[HandleSessionJoin] GetUserById error: %v", err)
		return
	}
	log.Printf("[HandleSessionJoin] Found user: %s", user.Name)

	err = h.DB.ActivateSessionParticipant(session, user)
	if err != nil {
		log.Printf("[HandleSessionJoin] ActivateSessionParticipant error: %v", err)
	}

	err = h.sendParticipantsList(session.ID)
	if err != nil {
		log.Printf("[HandleSessionJoin] sendParticipantsList error: %v", err)
	}
}

func (h *Handlers) HandleSessionLeave(session_id string, user_id string) {
	log.Printf("[HandleSessionLeave] start - session_id=%s, user_id=%s", session_id, user_id)
	session, err := h.DB.GetSession(session_id)
	if err != nil {
		log.Printf("[HandleSessionLeave] GetSession error: %v", err)
		return
	}
	log.Printf("[HandleSessionLeave] Found session with ID: %s", session.ID)

	user, err := h.DB.GetUserById(user_id)
	if err != nil {
		log.Printf("[HandleSessionLeave] GetUserById error: %v", err)
		return
	}
	log.Printf("[HandleSessionLeave] Found user: %s", user.Name)

	err = h.DB.DeactivateSessionParticipant(session, user)
	if err != nil {
		log.Printf("[HandleSessionLeave] DeactivateSessionParticipant error: %v", err)
	}

	err = h.sendParticipantsList(session.ID)
	if err != nil {
		log.Printf("[HandleSessionLeave] sendParticipantsList error: %v", err)
	}
}

func (h *Handlers) sendParticipantsList(session_id string) error {
	session, err := h.DB.GetSession(session_id)
	if err != nil {
		log.Printf("[sendParticipantsList] GetSession error: %v", err)
		return err
	}
	log.Printf("[sendParticipantsList] start - session.ID=%s", session.ID)

	log.Printf("[sendParticipantsList] Total participants: %d", len(session.Participants))
	for i, participant := range session.Participants {
		log.Printf("[sendParticipantsList] Participant %d: %s (IsOnline: %t)",
			i, participant.User.Name, participant.IsOnline)
	}

	online := organisms.GetOnline(session)
	log.Printf("[sendParticipantsList] Online participants: %d", len(online))

	var buf bytes.Buffer
	err = organisms.ParticipantsList(session).Render(context.Background(), &buf)
	if err != nil {
		log.Printf("[sendParticipantsList] Render error: %v", err)
		return err
	}

	html := fmt.Sprintf(`
    <div id="participants-list" hx-swap-oob="outerHTML class="w-full flex flex-row"">
      %s
      <button
        class="w-14 h-14 rounded-full bg-secondary-700 ring-2 ring-primary-400 hover:bg-primary-500 flex items-center justify-center text-xl font-bold border-2 border-primary-700 shadow -ml-2 transition-all duration-300 ease-in-out transform hover:scale-110 hover:rotate-12 hover:shadow-lg hover:ring-primary-300 active:scale-95 active:rotate-0 active:bg-primary-600 active:shadow-inner"
        title="Invite participants"
        hx-trigger="click"
        hx-target="#invite-popup"
        hx-get={"/session/%s/invite"}
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 transition-transform duration-300 group-hover:scale-110 group-active:scale-90">
          <path stroke-linecap="round" stroke-linejoin="round" d="M18 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0ZM3 19.235v-.11a6.375 6.375 0 0 1 12.75 0v.109A12.318 12.318 0 0 1 9.374 21c-2.331 0-4.512-.645-6.374-1.766Z"></path>
        </svg>
      </button>
    </div>
    `, buf.String(), session_id)

	log.Printf("[sendParticipantsList] Rendered HTML length: %d", len(html))

	sessionClientsMu.Lock()
	defer sessionClientsMu.Unlock()

	clients, ok := sessionClients[session.ID]
	if !ok {
		log.Printf("[sendParticipantsList] No clients map for session %s", session.ID)
		return nil
	}

	if len(clients) == 0 {
		log.Printf("[sendParticipantsList] No clients in session %s", session.ID)
		return nil
	}

	log.Printf("[sendParticipantsList] Sending to %d clients", len(clients))

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(html)); err != nil {
			log.Printf("[sendParticipantsList] Error writing to websocket: %v", err)
			delete(clients, conn)
			err = conn.Close()
			if err != nil {
				return err
			}
		} else {
			log.Printf("[sendParticipantsList] Successfully sent message to client")
		}
	}
	return nil
}

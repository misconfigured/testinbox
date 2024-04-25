package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var database *Database
var log = logrus.New()

type TemplateRenderer struct {
	templates *template.Template
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func init() {
	log.Out = os.Stdout

	env := os.Getenv("ENVIRONMENT")
	switch env {
	case "production", "staging":
		log.Formatter = &logrus.JSONFormatter{}
	default:
		log.Formatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}
	}

	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		log.Warnf("Invalid or no LOG_LEVEL provided; defaulting to 'info'. Error: %v", err)
		log.Level = logrus.InfoLevel
	} else {
		log.Level = level
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer() *TemplateRenderer {
	templates := template.New("").Funcs(template.FuncMap{
		"odd":           IsOdd,
		"truncateEmail": TruncateEmail,
		"add":           func(a, b int) int { return a + b },
		"sub":           func(a, b int) int { return a - b },
	})
	templates, err := templates.ParseGlob(filepath.Join("views", "*.html"))
	if err != nil {
		log.Fatal("Error loading templates:", err)
	}
	return &TemplateRenderer{
		templates: templates,
	}
}

func cacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=600")
		return next(c)
	}
}

func main() {
	e := echo.New()
	e.Renderer = NewTemplateRenderer()

	connectDB()
	database = &Database{Conn: db}

	e.Use(middleware.Gzip())

	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "Up!")
	})

	e.GET("/", homePage)
	e.GET("/inbox", fetchEmails)
	e.GET("/inbox/:id", fetchEmailDetails)
	e.GET("/api/suggestions/recipients", getRecipientSuggestions)
	e.POST("/webhook/sendgrid/receive", receiveSendGridWebhook)
	e.Group("/public", cacheControlMiddleware).Static("/", "public")
	e.GET("/ws", websocketHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

func receiveSendGridWebhook(c echo.Context) error {
	if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
		log.WithError(err).Error("Error parsing multipart form")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request, could not parse form"})
	}

	email := Email{
		Sender:          c.FormValue("from"),
		Recipient:       c.FormValue("to"),
		Subject:         c.FormValue("subject"),
		BodyHTML:        template.HTML(c.FormValue("html")),
		BodyText:        c.FormValue("text"),
		Headers:         c.FormValue("headers"),
		SenderIP:        c.FormValue("sender_ip"),
		SpamScore:       parseFloat(c.FormValue("spam_score")),
		AttachmentsInfo: c.FormValue("attachment-info"),
		ReceivedAt:      time.Now(),
		DKIM:            c.FormValue("dkim"),
		ContentIDs:      c.FormValue("content-ids"),
		Envelope:        c.FormValue("envelope"),
		Attachments:     parseInt(c.FormValue("attachments")),
		SpamReport:      c.FormValue("spam_report"),
		AttachmentInfo:  c.FormValue("attachment-info"),
		Charsets:        c.FormValue("charsets"),
		SPF:             c.FormValue("SPF"),
	}

	ctx := context.Background()
	if err := database.InsertEmail(ctx, email); err != nil {
		log.WithError(err).Error("Error inserting email into database")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("could not insert email: %v", err)})
	}

	emailData, err := json.Marshal(email)
	if err != nil {
		log.WithError(err).Error("Failed to marshal email data")
	}
	log.Printf("Broadcasting message: %s", emailData)
	broadcastMessage(websocket.TextMessage, emailData)

	log.Debug("Email saved successfully")
	return c.JSON(http.StatusOK, map[string]string{"status": "message saved", "id": fmt.Sprintf("%d", email.ID)})

}

func fetchEmails(c echo.Context) error {
	recipient := c.QueryParam("recipient")
	log.WithField("recipient", recipient).Debug("Fetching emails with filter")

	log.Debug("Fetching emails: Start")

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		log.WithFields(logrus.Fields{"error": err, "page": page}).Warn("Invalid page number, defaulting to page 1")
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	log.WithFields(logrus.Fields{"page": page, "limit": limit, "offset": offset}).Debug("Fetching emails")

	emails, total, err := database.GetEmails(limit, offset, recipient)
	if err != nil {
		log.WithError(err).Error("Failed to fetch emails")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch emails")
	}

	log.WithField("count", len(emails)).Debug("Fetched emails successfully")

	totalPages := (total + limit - 1) / limit
	pages := make([]int, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	if err := c.Render(http.StatusOK, "inbox.html", map[string]interface{}{
		"Emails":      emails,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"Pages":       pages,
	}); err != nil {
		log.WithError(err).Error("Failed to render email list")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render email list")
	}

	log.Println("Fetching emails: Completed successfully")
	return nil
}

func parseFloat(str string) float64 {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.WithError(err).Error("Error converting string to float")
		return 0.0
	}
	return val
}

func parseInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		log.WithError(err).Error("Error converting string to int")
		return 0
	}
	return val
}

func IsOdd(num int) bool {
	return num%2 != 0
}

func (e Email) ContentType() string {
	if e.BodyHTML != "" {
		return "HTML"
	} else if e.BodyText != "" {
		return "Text"
	}
	return "None"
}

func fetchEmailDetails(c echo.Context) error {
	emailId := c.Param("id")

	email, err := database.GetEmailByID(emailId)
	if err != nil {
		log.WithError(err).Error("Failed to fetch email")
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not fetch email")
	}

	if err := c.Render(http.StatusOK, "email.html", map[string]interface{}{
		"Email": email,
	}); err != nil {
		log.WithFields(logrus.Fields{
			"action": "renderEmailDetails",
			"error":  err,
		}).Error("Failed to render email details")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render email details")
	}
	return nil
}

func homePage(c echo.Context) error {
	log.Debug("Fetching latest 10 emails for homepage")
	emails, total, err := database.GetEmails(10, 0, "")
	if err != nil {
		log.WithError(err).Error("Failed to fetch emails for homepage")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch emails")
	}

	log.WithFields(logrus.Fields{
		"total_emails": total,
	}).Debug("Emails fetched for homepage")

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"Emails": emails,
		"Total":  total,
	})
}

func getRecipientSuggestions(c echo.Context) error {
	searchTerm := c.QueryParam("term")
	if len(searchTerm) < 2 {
		return c.JSON(http.StatusOK, []string{})
	}
	suggestions, err := database.GetRecipientSuggestions(searchTerm)
	if err != nil {
		log.WithError(err).Error("Failed to fetch recipient suggestions")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch suggestions")
	}

	return c.JSON(http.StatusOK, suggestions)
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func websocketHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Failed to upgrade WebSocket:", err)
		return err
	}
	defer conn.Close()

	// Register client
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			broadcastMessage(websocket.TextMessage, []byte("heartbeat"))
		}
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("Read error:", err)
			break
		}
	}
	return nil
}

func broadcastMessage(messageType int, message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	log.Printf("Broadcasting to %d clients", len(clients))
	for client := range clients {
		if err := client.WriteMessage(messageType, message); err != nil {
			log.Println("Failed to send message to a client:", err)
			continue
		}
	}
}

package main

import (
	"fmt"
	"html"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/testagent"
)

func runServer(cfg envConfig, reg *testagent.Registry) error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendString(renderDashboard(reg))
	})

	app.Get("/agents/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("invalid id")
		}
		a, err := reg.AgentByIndex(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}
		c.Type("html", "utf-8")
		return c.SendString(renderJournal(a))
	})

	app.Post("/agents/:id/start", func(c *fiber.Ctx) error {
		id, _ := c.ParamsInt("id")
		if err := reg.Start(id); err != nil {
			return c.Redirect("/?err=" + urlQuery(err.Error()))
		}
		return c.Redirect("/")
	})

	app.Post("/agents/:id/stop", func(c *fiber.Ctx) error {
		id, _ := c.ParamsInt("id")
		reg.Stop(id)
		return c.Redirect("/")
	})

	app.Post("/agents/start-all", func(c *fiber.Ctx) error {
		reg.StartAll()
		return c.Redirect("/")
	})

	app.Post("/agents/stop-all", func(c *fiber.Ctx) error {
		reg.StopAll()
		return c.Redirect("/")
	})

	app.Get("/api/agents", func(c *fiber.Ctx) error {
		return c.JSON(reg.Snapshots())
	})

	app.Get("/api/agents/:id/journal", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		a, err := reg.AgentByIndex(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(a.Journal().EntriesNewestFirst())
	})

	go func() {
		addr := ":" + cfg.Port
		logger.Info("test agents control UI listening", "addr", addr, "site", cfg.SiteURL)
		if err := app.Listen(addr); err != nil {
			logger.Error("control UI failed", "error", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("shutting down test agents")
	reg.StopAll()
	reg.WaitAll()
	return app.Shutdown()
}

func urlQuery(s string) string {
	return html.EscapeString(s)
}

func renderDashboard(reg *testagent.Registry) string {
	var b fmtBuilder
	b.write(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">`)
	b.write(`<meta http-equiv="refresh" content="5">`)
	b.write(`<title>Test Agents</title>`)
	b.write(`<style>
body{font-family:system-ui,sans-serif;margin:2rem;max-width:960px}
table{border-collapse:collapse;width:100%}
th,td{border:1px solid #ccc;padding:8px;text-align:left}
.status-running{color:green}.status-stalled{color:red}.status-stopped{color:#666}
.actions form{display:inline;margin:0 4px}
button{padding:4px 10px;cursor:pointer}
.toolbar{margin:1rem 0}
a{color:#06c}
</style></head><body>`)
	b.write(`<h1>Test Agents</h1>`)
	b.write(`<div class="toolbar">`)
	b.write(`<form method="post" action="/agents/start-all" style="display:inline">`)
	b.write(`<button type="submit">Start all</button></form> `)
	b.write(`<form method="post" action="/agents/stop-all" style="display:inline">`)
	b.write(`<button type="submit">Stop all</button></form>`)
	b.write(`</div>`)
	b.write(`<table><thead><tr>`)
	b.write(`<th>#</th><th>User</th><th>Persona</th><th>Status</th>`)
	b.write(`<th>Path</th><th>Last</th><th>Errors</th><th>Actions</th>`)
	b.write(`</tr></thead><tbody>`)

	for _, s := range reg.Snapshots() {
		cls := "status-" + string(s.Status)
		b.write("<tr>")
		b.writef("<td>%d</td>", s.Index)
		b.writef("<td>%s</td>", html.EscapeString(s.Username))
		b.writef("<td>%s</td>", html.EscapeString(s.Persona))
		b.writef(`<td class="%s">%s</td>`, cls, html.EscapeString(string(s.Status)))
		b.writef("<td>%s</td>", html.EscapeString(s.CurrentPath))
		b.writef("<td>%s</td>", html.EscapeString(s.LastAction))
		b.writef("<td>%d</td>", s.ErrorCount)
		b.write(`<td class="actions">`)
		b.writef(`<a href="/agents/%d">Journal</a> `, s.Index)
		if s.Status != testagent.StatusRunning {
			b.writef(`<form method="post" action="/agents/%d/start">`, s.Index)
			b.write(`<button type="submit">Start</button></form>`)
		} else {
			b.writef(`<form method="post" action="/agents/%d/stop">`, s.Index)
			b.write(`<button type="submit">Stop</button></form>`)
		}
		b.write(`</td></tr>`)
	}

	b.write(`</tbody></table></body></html>`)
	return b.String()
}

func renderJournal(a *testagent.Agent) string {
	snap := a.Snapshot()
	var b fmtBuilder
	b.write(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">`)
	b.write(`<meta http-equiv="refresh" content="5">`)
	b.writef(`<title>Agent %d Journal</title>`, snap.Index)
	b.write(`<style>
body{font-family:system-ui,sans-serif;margin:2rem;max-width:960px}
.entry{border-bottom:1px solid #eee;padding:8px 0}
.err{color:#c00}.phase{color:#666;font-size:0.85em}
</style></head><body>`)
	b.writef(`<p><a href="/">← Dashboard</a></p>`)
	b.writef(`<h1>Agent %d — %s</h1>`, snap.Index, html.EscapeString(snap.Username))
	b.writef(`<p>Persona: %s | Status: %s</p>`, html.EscapeString(snap.Persona), snap.Status)

	for _, e := range a.Journal().EntriesNewestFirst() {
		b.write(`<div class="entry">`)
		b.writef(`<span class="phase">%s %s</span> `,
			html.EscapeString(string(e.Phase)),
			html.EscapeString(e.Time.Format("15:04:05")))
		if e.URL != "" {
			b.writef(`<strong>%s</strong> `, html.EscapeString(e.URL))
		}
		if e.Action != "" {
			b.write(html.EscapeString(e.Action))
		}
		if e.Status != 0 {
			b.writef(` <em>%d</em>`, e.Status)
		}
		if e.Reasoning != "" {
			b.writef(`<div>%s</div>`, html.EscapeString(e.Reasoning))
		}
		if e.Error != "" {
			b.writef(`<div class="err">%s</div>`, html.EscapeString(e.Error))
		}
		b.write(`</div>`)
	}

	b.write(`</body></html>`)
	return b.String()
}

type fmtBuilder struct {
	s string
}

func (b *fmtBuilder) write(s string) { b.s += s }
func (b *fmtBuilder) writef(format string, args ...any) {
	b.s += fmt.Sprintf(format, args...)
}
func (b *fmtBuilder) String() string { return b.s }

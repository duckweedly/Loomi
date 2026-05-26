package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

type REPL struct {
	Client    *Client
	In        io.Reader
	Out       io.Writer
	Err       io.Writer
	ThreadID  string
	Mode      string
	Provider  string
	Model     string
	Persona   string
	Script    string
	LastRunID string
}

func (r *REPL) Run(ctx context.Context) error {
	in := r.In
	if in == nil {
		in = strings.NewReader("")
	}
	renderer := Renderer{Out: r.Out}
	scanner := bufio.NewScanner(in)
	_, _ = fmt.Fprintln(renderer.out(), "Loomi chat. /help for commands, /quit to exit.")
	for {
		_, _ = fmt.Fprint(renderer.out(), "loomi> ")
		if !scanner.Scan() {
			return scanner.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/") {
			done, err := r.handleCommand(ctx, line, renderer)
			if err != nil {
				_, _ = fmt.Fprintf(r.err(), "error: %v\n", err)
			}
			if done {
				return nil
			}
			continue
		}
		result, err := Runner{Client: r.client()}.Execute(ctx, RunOptions{
			ThreadID: r.ThreadID,
			Prompt:   line,
			Mode:     r.Mode,
			Provider: r.Provider,
			Model:    r.Model,
			Persona:  r.Persona,
			Script:   r.Script,
			OnEvent: func(event RunEvent) {
				_ = renderer.PrintEvent(event)
			},
			OnApproval: func(approval PendingApproval, event RunEvent) (string, error) {
				return promptApproval(scanner, renderer.out(), approval)
			},
		})
		if err != nil {
			_, _ = fmt.Fprintf(r.err(), "error: %v\n", err)
			continue
		}
		r.ThreadID = result.ThreadID
		r.LastRunID = result.RunID
		if err := renderer.PrintRunResult(result); err != nil {
			return err
		}
	}
}

func promptApproval(scanner *bufio.Scanner, out io.Writer, approval PendingApproval) (string, error) {
	toolName := approval.ToolName
	if toolName == "" {
		toolName = "tool"
	}
	for {
		if _, err := fmt.Fprintf(out, "approve %s %s? [a]pprove/[d]eny/[s]kip: ", toolName, approval.ToolCallID); err != nil {
			return "", err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "a", "approve", "y", "yes":
			return "approve", nil
		case "d", "deny", "n", "no":
			return "deny", nil
		case "s", "skip", "":
			return "skip", nil
		default:
			if _, err := fmt.Fprintln(out, "enter approve, deny, or skip"); err != nil {
				return "", err
			}
		}
	}
}

func (r *REPL) handleCommand(ctx context.Context, line string, renderer Renderer) (bool, error) {
	fields := strings.Fields(line)
	cmd := fields[0]
	args := fields[1:]
	switch cmd {
	case "/quit", "/exit":
		return true, nil
	case "/help":
		_, err := fmt.Fprintln(renderer.out(), "/status\n/thread\n/new\n/model <provider-id-or-model>\n/persona <id-or-slug>\n/tools [group]\n/approvals [run-id]\n/events [compact] [run-id]\n/quit")
		return false, err
	case "/status":
		return false, renderer.PrintStatus(r.client())
	case "/thread":
		threadID := strings.TrimSpace(r.ThreadID)
		if threadID == "" {
			threadID = "(new thread)"
		}
		_, err := fmt.Fprintf(renderer.out(), "%s\n", threadID)
		return false, err
	case "/new":
		r.ThreadID = ""
		_, err := fmt.Fprintln(renderer.out(), "new thread")
		return false, err
	case "/model":
		if len(args) != 1 {
			return false, errors.New("usage: /model <provider-id-or-model>")
		}
		return false, r.setModel(ctx, args[0])
	case "/persona":
		if len(args) != 1 {
			return false, errors.New("usage: /persona <id-or-slug>")
		}
		return false, r.setPersona(ctx, args[0])
	case "/tools":
		return false, r.printTools(ctx, args, renderer)
	case "/approvals":
		return false, r.printApprovals(ctx, args, renderer)
	case "/events":
		return false, r.printEvents(ctx, args, renderer)
	default:
		return false, fmt.Errorf("unknown command %s", cmd)
	}
}

func (r *REPL) printTools(ctx context.Context, args []string, renderer Renderer) error {
	if len(args) > 1 {
		return errors.New("usage: /tools [group]")
	}
	tools, err := r.client().ListTools(ctx)
	if err != nil {
		return err
	}
	if len(args) == 1 {
		group := strings.TrimSpace(args[0])
		filtered := make([]ToolCatalogEntry, 0, len(tools))
		for _, tool := range tools {
			if tool.Group == group {
				filtered = append(filtered, tool)
			}
		}
		tools = filtered
	}
	if len(tools) == 0 {
		_, err := fmt.Fprintln(renderer.out(), "no tools")
		return err
	}
	return renderer.PrintTools(tools)
}

func (r *REPL) printApprovals(ctx context.Context, args []string, renderer Renderer) error {
	if len(args) > 1 {
		return errors.New("usage: /approvals [run-id]")
	}
	runID := r.runIDArg(args)
	if runID == "" {
		_, err := fmt.Fprintln(renderer.out(), "no run yet")
		return err
	}
	events, err := r.client().ListEvents(ctx, runID, 0)
	if err != nil {
		return err
	}
	pending := PendingApprovalEvents(events)
	if len(pending) == 0 {
		_, err := fmt.Fprintln(renderer.out(), "no pending approvals")
		return err
	}
	return renderer.PrintApprovals(pending)
}

func (r *REPL) printEvents(ctx context.Context, args []string, renderer Renderer) error {
	compact := false
	if len(args) > 0 && args[0] == "compact" {
		compact = true
		args = args[1:]
	}
	if len(args) > 1 {
		return errors.New("usage: /events [compact] [run-id]")
	}
	runID := r.runIDArg(args)
	if runID == "" {
		_, err := fmt.Fprintln(renderer.out(), "no run yet")
		return err
	}
	events, err := r.client().ListEvents(ctx, runID, 0)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		_, err := fmt.Fprintln(renderer.out(), "no events")
		return err
	}
	for _, event := range events {
		if compact {
			if err := renderer.PrintEventCompact(event); err != nil {
				return err
			}
			continue
		}
		if err := renderer.PrintEvent(event); err != nil {
			return err
		}
	}
	return nil
}

func (r *REPL) runIDArg(args []string) string {
	if len(args) == 1 {
		return strings.TrimSpace(args[0])
	}
	return strings.TrimSpace(r.LastRunID)
}

func (r *REPL) setModel(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	providers, err := r.client().ListModelProviders(ctx)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if provider.ID == value || provider.Model == value {
			r.Provider = provider.ID
			r.Model = provider.Model
			return nil
		}
	}
	return fmt.Errorf("model provider %q not found", value)
}

func (r *REPL) setPersona(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	personas, err := r.client().ListPersonas(ctx)
	if err != nil {
		return err
	}
	for _, persona := range personas {
		if persona.ID == value || persona.Slug == value {
			r.Persona = persona.ID
			return nil
		}
	}
	return fmt.Errorf("persona %q not found", value)
}

func (r *REPL) client() *Client {
	if r.Client != nil {
		return r.Client
	}
	return NewClient("")
}

func (r *REPL) err() io.Writer {
	if r.Err != nil {
		return r.Err
	}
	return io.Discard
}

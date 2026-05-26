package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sheridiany/loomi/internal/cli"
)

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if ee, ok := err.(exitError); ok {
			os.Exit(ee.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	return runWithIO(args, os.Stdin, stdout, stderr)
}

func runWithIO(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stderr)
		return exitError{code: 2}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch args[0] {
	case "help":
		return cmdHelp(args[1:], stdout)
	case "status":
		return cmdStatus(ctx, args[1:], stdout)
	case "config":
		return cmdConfig(args[1:], stdout)
	case "tools":
		return cmdTools(ctx, args[1:], stdout)
	case "run":
		return cmdRun(ctx, args[1:], stdin, stdout)
	case "chat":
		return cmdChat(ctx, args[1:], stdin, stdout, stderr)
	case "sessions":
		return cmdSessions(ctx, args[1:], stdin, stdout, stderr)
	case "models":
		return cmdModels(ctx, args[1:], stdout)
	case "personas":
		return cmdPersonas(ctx, args[1:], stdout)
	case "events":
		return cmdEvents(ctx, args[1:], stdout)
	case "approvals":
		return cmdApprovals(ctx, args[1:], stdout)
	default:
		printUsage(stderr)
		return exitError{code: 2}
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `usage: loomi <command> [flags]

commands:
  help [topic]                    show command help
  status                         check local Loomi API
  config show                    show resolved local CLI defaults
  config set <key> <value>       persist a local CLI default
  config unset <key>             remove a local CLI default
  chat                           open an interactive Loomi chat shell with inline approvals
  sessions list                  list threads
  sessions resume <thread-id>    resume a thread in the chat shell
  models list                    list model providers
  personas list                  list personas
  tools list                     list runtime tools
  run [--compact] [--interactive-approvals] <prompt>
                                 create/send/start a run and stream events
  events tail [--tools-only] <run-id>
                                 stream run events
  approvals list <run-id>        list pending approval events for a run
  approvals follow <run-id>      stream approval-focused notices
  approvals approve [--follow] <thread-id> <run-id> <tool-call-id>
  approvals deny [--follow] <thread-id> <run-id> <tool-call-id>`)
}

func cmdHelp(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}
	switch strings.TrimSpace(args[0]) {
	case "run":
		_, err := fmt.Fprintln(stdout, `usage: loomi run [flags] <prompt>

flags:
  --host <url>                  Loomi API host
  --thread <id>                 use an existing thread
  --mode <chat|work>            thread mode, default from config
  --provider <id>               model provider id
  --model <model>               model override
  --persona <id>                persona id
  --script <name>               local simulated script
  --prompt-file <path|->        read prompt from a file or stdin
  --timeout <duration>          run timeout, for example 2m
  --compact                     show shorter live event transcript
  --interactive-approvals       approve/deny/skip tool calls inline
  --output text|json|stream-json`)
		return err
	case "tools":
		_, err := fmt.Fprintln(stdout, `usage: loomi tools list [flags]

flags:
  --host <url>          Loomi API host
  --group <name>        show one tool group
  --enabled-only        show only enabled tools
  --flat                print legacy tabular rows
  --output text|json`)
		return err
	case "approvals":
		_, err := fmt.Fprintln(stdout, `usage: loomi approvals <command>

commands:
  list <run-id>
  follow <run-id>
  approve [--follow] <thread-id> <run-id> <tool-call-id>
  deny [--follow] <thread-id> <run-id> <tool-call-id>`)
		return err
	case "events":
		_, err := fmt.Fprintln(stdout, `usage: loomi events tail [flags] <run-id>

flags:
  --host <url>          Loomi API host
  --after <sequence>    resume after event sequence
  --tools-only          show only tool call events
  --compact             show compact one-line event summaries
  --output text|json`)
		return err
	case "config":
		_, err := fmt.Fprintln(stdout, `usage: loomi config <command>

commands:
  show [--output text|json]
  set <host|mode|provider|model|persona|script> <value>
  unset <host|mode|provider|model|persona|script>`)
		return err
	default:
		return fmt.Errorf("unknown help topic %s", args[0])
	}
}

func cmdStatus(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	if err := fs.Parse(args); err != nil {
		return err
	}
	client := cli.NewClient(*host)
	if err := client.CheckReady(ctx); err != nil {
		return err
	}
	return cli.Renderer{Out: stdout}.PrintStatus(client)
}

func cmdConfig(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi config show|set|unset")
	}
	switch args[0] {
	case "show":
		return cmdConfigShow(args[1:], stdout)
	case "set":
		return cmdConfigSet(args[1:], stdout)
	case "unset":
		return cmdConfigUnset(args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi config show|set|unset")
	}
}

func cmdConfigShow(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("config show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(cfg)
	}
	_, err = fmt.Fprintf(stdout, "path\t%s\nfound\t%v\nhost\t%s\nmode\t%s\nprovider\t%s\nmodel\t%s\npersona\t%s\nscript\t%s\n", cfg.Path, cfg.Found, cfg.Host, cfg.Mode, cfg.Provider, cfg.Model, cfg.Persona, cfg.Script)
	return err
}

func cmdConfigSet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("config set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: loomi config set <host|mode|provider|model|persona|script> <value>")
	}
	cfg, err := cli.LoadConfigFileFromEnv()
	if err != nil {
		return err
	}
	if err := cli.SetConfigValue(&cfg, fs.Arg(0), fs.Arg(1)); err != nil {
		return err
	}
	if err := cli.SaveConfigFile(cfg); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "set %s in %s\n", fs.Arg(0), cfg.Path)
	return err
}

func cmdConfigUnset(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("config unset", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi config unset <host|mode|provider|model|persona|script>")
	}
	cfg, err := cli.LoadConfigFileFromEnv()
	if err != nil {
		return err
	}
	if err := cli.UnsetConfigValue(&cfg, fs.Arg(0)); err != nil {
		return err
	}
	if err := cli.SaveConfigFile(cfg); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "unset %s in %s\n", fs.Arg(0), cfg.Path)
	return err
}

func cmdTools(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("usage: loomi tools list [--host %s]", cli.DefaultBaseURL)
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("tools list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	group := fs.String("group", "", "tool group filter")
	enabledOnly := fs.Bool("enabled-only", false, "show only enabled tools")
	flat := fs.Bool("flat", false, "print flat tabular rows")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	tools, err := cli.NewClient(*host).ListTools(ctx)
	if err != nil {
		return err
	}
	tools = filterTools(tools, *group, *enabledOnly)
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(tools)
	}
	if *flat {
		return renderer.PrintToolsFlat(tools)
	}
	return renderer.PrintTools(tools)
}

func filterTools(tools []cli.ToolCatalogEntry, group string, enabledOnly bool) []cli.ToolCatalogEntry {
	group = strings.TrimSpace(group)
	if group == "" && !enabledOnly {
		return tools
	}
	filtered := make([]cli.ToolCatalogEntry, 0, len(tools))
	for _, tool := range tools {
		if group != "" && tool.Group != group {
			continue
		}
		if enabledOnly && !tool.Enabled {
			continue
		}
		filtered = append(filtered, tool)
	}
	return filtered
}

func cmdRun(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	mode := fs.String("mode", cfg.Mode, "thread mode")
	threadID := fs.String("thread", "", "existing thread id")
	provider := fs.String("provider", cfg.Provider, "provider id")
	model := fs.String("model", cfg.Model, "model override")
	persona := fs.String("persona", cfg.Persona, "persona id")
	script := fs.String("script", cfg.Script, "local simulated script")
	promptFile := fs.String("prompt-file", "", "prompt file path, or - for stdin")
	output := fs.String("output", "text", "text, json, or stream-json")
	timeout := fs.Duration("timeout", 0, "run timeout, for example 2m")
	compact := fs.Bool("compact", false, "show shorter live event transcript")
	interactiveApprovals := fs.Bool("interactive-approvals", false, "prompt to approve or deny tool calls during the run")
	if err := fs.Parse(args); err != nil {
		return err
	}
	prompt, err := resolvePrompt(stdin, fs.Args(), *promptFile)
	if err != nil {
		return err
	}
	if prompt == "" {
		return fmt.Errorf("usage: loomi run <prompt|--prompt-file path|--prompt-file ->")
	}
	if *output != "text" && *output != "json" && *output != "stream-json" {
		return fmt.Errorf("--output must be text, json, or stream-json")
	}
	if *interactiveApprovals && *output != "text" {
		return fmt.Errorf("--interactive-approvals requires --output text")
	}
	if *compact && *output != "text" {
		return fmt.Errorf("--compact requires --output text")
	}
	if *interactiveApprovals && *promptFile == "-" {
		return fmt.Errorf("--interactive-approvals cannot share stdin with --prompt-file -")
	}
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	renderer := cli.Renderer{Out: stdout}
	onEvent := func(event cli.RunEvent) {
		if *compact {
			_ = renderer.PrintEventCompact(event)
			return
		}
		_ = renderer.PrintEvent(event)
	}
	if *output == "json" {
		onEvent = nil
	}
	if *output == "stream-json" {
		onEvent = func(event cli.RunEvent) {
			_ = renderer.PrintJSONLine(event)
		}
	}
	var onApproval func(cli.PendingApproval, cli.RunEvent) (string, error)
	if *interactiveApprovals {
		reader := bufio.NewReader(stdin)
		onApproval = func(approval cli.PendingApproval, event cli.RunEvent) (string, error) {
			toolName := approval.ToolName
			if toolName == "" {
				toolName = "tool"
			}
			for {
				if _, err := fmt.Fprintf(stdout, "approve %s %s? [a]pprove/[d]eny/[s]kip: ", toolName, approval.ToolCallID); err != nil {
					return "", err
				}
				line, err := reader.ReadString('\n')
				if err != nil && strings.TrimSpace(line) == "" {
					return "", err
				}
				switch strings.ToLower(strings.TrimSpace(line)) {
				case "a", "approve", "y", "yes":
					return "approve", nil
				case "d", "deny", "n", "no":
					return "deny", nil
				case "s", "skip", "":
					return "skip", nil
				default:
					if _, err := fmt.Fprintln(stdout, "enter approve, deny, or skip"); err != nil {
						return "", err
					}
				}
			}
		}
	}
	result, err := cli.Runner{Client: cli.NewClient(*host)}.Execute(ctx, cli.RunOptions{
		ThreadID:   *threadID,
		Prompt:     prompt,
		Mode:       *mode,
		Provider:   *provider,
		Model:      *model,
		Persona:    *persona,
		Script:     *script,
		OnEvent:    onEvent,
		OnApproval: onApproval,
	})
	if err != nil {
		return err
	}
	if *output == "json" {
		return renderer.PrintJSON(result)
	}
	if *output == "stream-json" {
		return renderer.PrintJSONLine(result)
	}
	return renderer.PrintRunResult(result)
}

func cmdChat(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("chat", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	mode := fs.String("mode", cfg.Mode, "thread mode")
	threadID := fs.String("thread", "", "existing thread id")
	provider := fs.String("provider", cfg.Provider, "provider id")
	model := fs.String("model", cfg.Model, "model override")
	persona := fs.String("persona", cfg.Persona, "persona id")
	script := fs.String("script", cfg.Script, "local simulated script")
	timeout := fs.Duration("timeout", 0, "chat command timeout, for example 30m")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	return (&cli.REPL{
		Client:   cli.NewClient(*host),
		In:       stdin,
		Out:      stdout,
		Err:      stderr,
		ThreadID: *threadID,
		Mode:     *mode,
		Provider: *provider,
		Model:    *model,
		Persona:  *persona,
		Script:   *script,
	}).Run(ctx)
}

func cmdSessions(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi sessions list|resume")
	}
	switch args[0] {
	case "list":
		return cmdSessionsList(ctx, args[1:], stdout)
	case "resume":
		return cmdSessionsResume(ctx, args[1:], stdin, stdout, stderr)
	default:
		return fmt.Errorf("usage: loomi sessions list|resume")
	}
}

func cmdSessionsList(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("sessions list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	threads, err := cli.NewClient(*host).ListThreads(ctx)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(threads)
	}
	return renderer.PrintThreads(threads)
}

func cmdSessionsResume(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("sessions resume", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	timeout := fs.Duration("timeout", 0, "chat command timeout, for example 30m")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi sessions resume <thread-id>")
	}
	chatArgs := []string{"--host", *host, "--thread", fs.Arg(0)}
	if *timeout > 0 {
		chatArgs = append(chatArgs, "--timeout", timeout.String())
	}
	return cmdChat(ctx, chatArgs, stdin, stdout, stderr)
}

func cmdModels(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("usage: loomi models list [--host %s]", cli.DefaultBaseURL)
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("models list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	providers, err := cli.NewClient(*host).ListModelProviders(ctx)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(providers)
	}
	return renderer.PrintModelProviders(providers)
}

func cmdPersonas(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("usage: loomi personas list [--host %s]", cli.DefaultBaseURL)
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("personas list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	personas, err := cli.NewClient(*host).ListPersonas(ctx)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(personas)
	}
	return renderer.PrintPersonas(personas)
}

func cmdEvents(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "tail" {
		return fmt.Errorf("usage: loomi events tail <run-id>")
	}
	fs := flag.NewFlagSet("events tail", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cli.DefaultBaseURL, "Loomi API host")
	after := fs.Int("after", 0, "after sequence")
	output := fs.String("output", "text", "text or json")
	toolsOnly := fs.Bool("tools-only", false, "show only tool call events")
	compact := fs.Bool("compact", false, "show compact one-line event summaries")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi events tail <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	renderer := cli.Renderer{Out: stdout}
	return cli.NewClient(*host).StreamEvents(ctx, fs.Arg(0), *after, func(event cli.RunEvent) {
		if *toolsOnly && !cli.IsToolEvent(event) {
			return
		}
		if *output == "json" {
			_ = renderer.PrintJSONLine(event)
			return
		}
		if *compact {
			_ = renderer.PrintEventCompact(event)
			return
		}
		_ = renderer.PrintEvent(event)
	})
}

func cmdApprovals(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi approvals list|approve|deny")
	}
	switch args[0] {
	case "list":
		return cmdApprovalsList(ctx, args[1:], stdout)
	case "follow":
		return cmdApprovalsFollow(ctx, args[1:], stdout)
	case "approve", "deny":
		return cmdApprovalDecision(ctx, args[0], args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi approvals list|follow|approve|deny")
	}
}

func cmdApprovalsList(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("approvals list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi approvals list <run-id>")
	}
	events, err := cli.NewClient(*host).ListEvents(ctx, fs.Arg(0), 0)
	if err != nil {
		return err
	}
	return cli.Renderer{Out: stdout}.PrintApprovals(cli.PendingApprovalEvents(events))
}

func cmdApprovalsFollow(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("approvals follow", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	after := fs.Int("after", 0, "after sequence")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi approvals follow <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	renderer := cli.Renderer{Out: stdout}
	return cli.NewClient(*host).StreamEvents(ctx, fs.Arg(0), *after, func(event cli.RunEvent) {
		if *output == "json" {
			if event.Type == "tool_call_approval_required" || strings.HasPrefix(event.Type, "tool_call_") {
				_ = renderer.PrintJSONLine(event)
			}
			return
		}
		_ = renderer.PrintApprovalNotice(event)
	})
}

func cmdApprovalDecision(ctx context.Context, action string, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("approvals "+action, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	follow := fs.Bool("follow", false, "stream run events after the decision")
	after := fs.Int("after", -1, "event sequence to resume after; defaults to current last sequence when --follow is set")
	output := fs.String("output", "text", "text or json when --follow is set")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 3 {
		return fmt.Errorf("usage: loomi approvals %s [--follow] <thread-id> <run-id> <tool-call-id>", action)
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	client := cli.NewClient(*host)
	resumeAfter := *after
	if *follow && resumeAfter < 0 {
		events, err := client.ListEvents(ctx, fs.Arg(1), 0)
		if err != nil {
			return err
		}
		resumeAfter = maxEventSequence(events)
	}
	call, err := client.DecideToolCall(ctx, fs.Arg(0), fs.Arg(1), fs.Arg(2), action)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "%s %s %s\n", call.ToolCallID, call.ApprovalStatus, call.ExecutionStatus); err != nil {
		return err
	}
	if !*follow {
		return nil
	}
	renderer := cli.Renderer{Out: stdout}
	return client.StreamEvents(ctx, fs.Arg(1), resumeAfter, func(event cli.RunEvent) {
		if *output == "json" {
			_ = renderer.PrintJSONLine(event)
			return
		}
		_ = renderer.PrintEvent(event)
	})
}

func maxEventSequence(events []cli.RunEvent) int {
	max := 0
	for _, event := range events {
		if event.Sequence > max {
			max = event.Sequence
		}
	}
	return max
}

func resolvePrompt(stdin io.Reader, args []string, promptFile string) (string, error) {
	promptFile = strings.TrimSpace(promptFile)
	if promptFile != "" {
		var raw []byte
		var err error
		if promptFile == "-" {
			raw, err = io.ReadAll(stdin)
		} else {
			raw, err = os.ReadFile(promptFile)
		}
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(raw)), nil
	}
	return strings.TrimSpace(strings.Join(args, " ")), nil
}

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
	case "version":
		return cmdVersion(args[1:], stdout)
	case "completion":
		return cmdCompletion(args[1:], stdout)
	case "doctor":
		return cmdDoctor(ctx, args[1:], stdout)
	case "status":
		return cmdStatus(ctx, args[1:], stdout)
	case "config":
		return cmdConfig(args[1:], stdout)
	case "tools":
		return cmdTools(ctx, args[1:], stdout)
	case "mcp":
		return cmdMCP(ctx, args[1:], stdout)
	case "lsp":
		return cmdLSP(ctx, args[1:], stdout)
	case "artifacts":
		return cmdArtifacts(ctx, args[1:], stdout)
	case "memory":
		return cmdMemory(ctx, args[1:], stdout)
	case "agent":
		return cmdAgent(ctx, args[1:], stdout)
	case "browser":
		return cmdBrowser(ctx, args[1:], stdout)
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
	case "runs":
		return cmdRuns(ctx, args[1:], stdout)
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
  version                        show CLI version
  completion <bash|zsh|fish>      print shell completion script
  doctor                         check API, config, provider, and tools
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
  mcp servers                    list safe MCP server status
  lsp tools                      list LSP tool catalog entries
  artifacts list <thread-id>     list thread artifacts
  artifacts read <thread-id> <artifact-id>
                                 read one artifact excerpt
  memory list|search|show|audit  inspect memory safely
  agent tasks <thread-id>         list coordination-only agent tasks
  browser tools|events            inspect browser tool surface and events
  run [--compact] [--interactive-approvals] <prompt>
                                 create/send/start a run and stream events
  runs status <run-id>           show run status
  runs stop <run-id>             stop a queued or running run
  runs attach <run-id>           replay current run events and continue streaming
  runs follow <run-id>           stream only new run events by default
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
	case "completion":
		_, err := fmt.Fprintln(stdout, `usage: loomi completion <bash|zsh|fish>`)
		return err
	case "version":
		_, err := fmt.Fprintln(stdout, `usage: loomi version [flags]

flags:
  --output text|json`)
		return err
	case "doctor":
		_, err := fmt.Fprintln(stdout, `usage: loomi doctor [flags]

flags:
  --host <url>          Loomi API host
  --output text|json`)
		return err
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
	case "mcp":
		_, err := fmt.Fprintln(stdout, `usage: loomi mcp servers [flags]

flags:
  --host <url>          Loomi API host
  --output text|json`)
		return err
	case "lsp":
		_, err := fmt.Fprintln(stdout, `usage: loomi lsp tools [flags]

flags:
  --host <url>          Loomi API host
  --output text|json`)
		return err
	case "artifacts":
		_, err := fmt.Fprintln(stdout, `usage: loomi artifacts <command>

commands:
  list [--limit n] <thread-id>
  read [--max-bytes n] <thread-id> <artifact-id>`)
		return err
	case "memory":
		_, err := fmt.Fprintln(stdout, `usage: loomi memory <command>

commands:
  list [--limit n] [--scope-type type --scope-id id]
  search <query> [--limit n] [--scope-type type --scope-id id]
  show <entry-id> [--scope-type type --scope-id id]
  audit [--limit n] [--thread-id id] [--source-run-id id]`)
		return err
	case "agent":
		_, err := fmt.Fprintln(stdout, `usage: loomi agent <command>

commands:
  tasks [--limit n] <thread-id>
  tools`)
		return err
	case "browser":
		_, err := fmt.Fprintln(stdout, `usage: loomi browser <command>

commands:
  tools
  events [--compact] <run-id>`)
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
	case "runs":
		_, err := fmt.Fprintln(stdout, `usage: loomi runs <command>

commands:
  status <run-id>
  stop <run-id>
  attach [--after <sequence>] [--compact] [--tools-only] <run-id>
  follow [--after <sequence>] [--compact] [--tools-only] <run-id>`)
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

func cmdVersion(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	info := currentVersion()
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(info)
	}
	_, err := fmt.Fprintf(stdout, "loomi %s\ncommit %s\ndate %s\n", info.Version, info.Commit, info.Date)
	return err
}

func cmdDoctor(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	cfg.Host = *host
	report := cli.RunDoctor(ctx, cli.NewClient(*host), cfg)
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		if err := renderer.PrintJSON(report); err != nil {
			return err
		}
	} else if err := renderer.PrintDoctor(report); err != nil {
		return err
	}
	if !report.OK {
		return exitError{code: 1}
	}
	return nil
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

func cmdMCP(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "servers" {
		return fmt.Errorf("usage: loomi mcp servers")
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("mcp servers", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	servers, err := cli.NewClient(*host).ListMCPServers(ctx)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(servers)
	}
	return renderer.PrintMCPServers(servers)
}

func cmdLSP(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "tools" {
		return fmt.Errorf("usage: loomi lsp tools")
	}
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("lsp tools", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	tools, err := cli.NewClient(*host).ListTools(ctx)
	if err != nil {
		return err
	}
	lspTools := make([]cli.ToolCatalogEntry, 0, len(tools))
	for _, tool := range tools {
		if strings.TrimSpace(tool.Group) == "lsp" || strings.HasPrefix(tool.Name, "lsp.") {
			lspTools = append(lspTools, tool)
		}
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(lspTools)
	}
	return renderer.PrintTools(lspTools)
}

func cmdArtifacts(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi artifacts list|read")
	}
	switch args[0] {
	case "list":
		return cmdArtifactsList(ctx, args[1:], stdout)
	case "read":
		return cmdArtifactsRead(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi artifacts list|read")
	}
}

func cmdArtifactsList(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("artifacts list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	limit := fs.Int("limit", 20, "artifact limit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi artifacts list <thread-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	artifacts, err := cli.NewClient(*host).ListArtifacts(ctx, fs.Arg(0), *limit)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(artifacts)
	}
	return renderer.PrintArtifacts(artifacts)
}

func cmdArtifactsRead(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("artifacts read", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	maxBytes := fs.Int("max-bytes", 4096, "max excerpt bytes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: loomi artifacts read <thread-id> <artifact-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	artifact, err := cli.NewClient(*host).ReadArtifact(ctx, fs.Arg(0), fs.Arg(1), *maxBytes)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(artifact)
	}
	return renderer.PrintArtifact(artifact)
}

func cmdMemory(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi memory list|search|show|audit")
	}
	switch args[0] {
	case "list":
		return cmdMemoryList(ctx, args[1:], stdout)
	case "search":
		return cmdMemorySearch(ctx, args[1:], stdout)
	case "show":
		return cmdMemoryShow(ctx, args[1:], stdout)
	case "audit":
		return cmdMemoryAudit(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi memory list|search|show|audit")
	}
}

func cmdMemoryList(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("memory list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	filters := memoryFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	items, err := cli.NewClient(*host).ListMemory(ctx, filters())
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(items)
	}
	return renderer.PrintMemoryItems(items)
}

func cmdMemorySearch(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("memory search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	filters := memoryFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loomi memory search <query>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	input := filters()
	input.Query = strings.TrimSpace(strings.Join(fs.Args(), " "))
	items, err := cli.NewClient(*host).SearchMemory(ctx, input)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(items)
	}
	return renderer.PrintMemoryItems(items)
}

func cmdMemoryShow(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("memory show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	filters := memoryFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi memory show <entry-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	item, err := cli.NewClient(*host).GetMemory(ctx, fs.Arg(0), filters())
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(item)
	}
	return renderer.PrintMemoryItem(item)
}

func cmdMemoryAudit(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("memory audit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	eventType := fs.String("event-type", "", "memory audit event type")
	threadID := fs.String("thread-id", "", "thread id")
	filters := memoryFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	input := filters()
	if input.SourceThreadID == "" {
		input.SourceThreadID = strings.TrimSpace(*threadID)
	}
	items, err := cli.NewClient(*host).ListMemoryAudit(ctx, input, *eventType)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(items)
	}
	return renderer.PrintMemoryAudit(items)
}

func cmdAgent(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi agent tasks|tools")
	}
	switch args[0] {
	case "tasks":
		return cmdAgentTasks(ctx, args[1:], stdout)
	case "tools":
		return cmdToolGroup(ctx, "agent", args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi agent tasks|tools")
	}
}

func cmdAgentTasks(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("agent tasks", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	limit := fs.Int("limit", 20, "task limit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi agent tasks <thread-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	tasks, err := cli.NewClient(*host).ListAgentTasks(ctx, fs.Arg(0), *limit)
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(tasks)
	}
	return renderer.PrintAgentTasks(tasks)
}

func cmdBrowser(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi browser tools|events")
	}
	switch args[0] {
	case "tools":
		return cmdToolGroup(ctx, "browser", args[1:], stdout)
	case "events":
		return cmdBrowserEvents(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi browser tools|events")
	}
}

func cmdBrowserEvents(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("browser events", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	after := fs.Int("after", 0, "after sequence")
	output := fs.String("output", "text", "text or json")
	compact := fs.Bool("compact", false, "show compact one-line event summaries")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi browser events <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	renderer := cli.Renderer{Out: stdout}
	return cli.NewClient(*host).StreamEvents(ctx, fs.Arg(0), *after, func(event cli.RunEvent) {
		if !strings.HasPrefix(toolNameFromEvent(event), "browser.") {
			return
		}
		_ = printStreamedEvent(renderer, event, *output, *compact, false)
	})
}

func cmdToolGroup(ctx context.Context, group string, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(group+" tools", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	tools, err := cli.NewClient(*host).ListTools(ctx)
	if err != nil {
		return err
	}
	filtered := filterTools(tools, group, false)
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(filtered)
	}
	return renderer.PrintTools(filtered)
}

func toolNameFromEvent(event cli.RunEvent) string {
	if event.Metadata == nil {
		return ""
	}
	if value, ok := event.Metadata["tool_name"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func memoryFlags(fs *flag.FlagSet) func() cli.MemoryFilters {
	limit := fs.Int("limit", 20, "result limit")
	scopeType := fs.String("scope-type", "", "memory scope type")
	scopeID := fs.String("scope-id", "", "memory scope id")
	sourceThreadID := fs.String("source-thread-id", "", "source thread id")
	sourceRunID := fs.String("source-run-id", "", "source run id")
	sourceType := fs.String("source-type", "", "source type")
	includeTombstoned := fs.Bool("include-tombstoned", false, "include tombstoned memory")
	return func() cli.MemoryFilters {
		return cli.MemoryFilters{
			Limit:             *limit,
			ScopeType:         strings.TrimSpace(*scopeType),
			ScopeID:           strings.TrimSpace(*scopeID),
			SourceThreadID:    strings.TrimSpace(*sourceThreadID),
			SourceRunID:       strings.TrimSpace(*sourceRunID),
			SourceType:        strings.TrimSpace(*sourceType),
			IncludeTombstoned: *includeTombstoned,
		}
	}
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
		_ = printStreamedEvent(renderer, event, *output, *compact, *toolsOnly)
	})
}

func printStreamedEvent(renderer cli.Renderer, event cli.RunEvent, output string, compact bool, toolsOnly bool) error {
	if toolsOnly && !cli.IsToolEvent(event) {
		return nil
	}
	if output == "json" {
		return renderer.PrintJSONLine(event)
	}
	if compact {
		return renderer.PrintEventCompact(event)
	}
	return renderer.PrintEvent(event)
}

func cmdRuns(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: loomi runs status|stop|attach|follow <run-id>")
	}
	switch args[0] {
	case "status":
		return cmdRunsStatus(ctx, args[1:], stdout)
	case "stop":
		return cmdRunsStop(ctx, args[1:], stdout)
	case "attach":
		return cmdRunsAttach(ctx, args[1:], stdout)
	case "follow":
		return cmdRunsFollow(ctx, args[1:], stdout)
	default:
		return fmt.Errorf("usage: loomi runs status|stop|attach|follow <run-id>")
	}
}

func cmdRunsStatus(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("runs status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi runs status <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	run, err := cli.NewClient(*host).GetRun(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(run)
	}
	return renderer.PrintRun(run)
}

func cmdRunsStop(ctx context.Context, args []string, stdout io.Writer) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("runs stop", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	output := fs.String("output", "text", "text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: loomi runs stop <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	result, err := cli.NewClient(*host).StopRun(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	renderer := cli.Renderer{Out: stdout}
	if *output == "json" {
		return renderer.PrintJSON(result)
	}
	return renderer.PrintStopRun(result)
}

func cmdRunsAttach(ctx context.Context, args []string, stdout io.Writer) error {
	return cmdRunsStream(ctx, args, stdout, true)
}

func cmdRunsFollow(ctx context.Context, args []string, stdout io.Writer) error {
	return cmdRunsStream(ctx, args, stdout, false)
}

func cmdRunsStream(ctx context.Context, args []string, stdout io.Writer, replay bool) error {
	cfg, err := cli.LoadConfigFromEnv()
	if err != nil {
		return err
	}
	name := "runs follow"
	defaultAfter := -1
	if replay {
		name = "runs attach"
		defaultAfter = 0
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "Loomi API host")
	after := fs.Int("after", defaultAfter, "event sequence to resume after")
	output := fs.String("output", "text", "text or json")
	toolsOnly := fs.Bool("tools-only", false, "show only tool call events")
	compact := fs.Bool("compact", false, "show compact one-line event summaries")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		if replay {
			return fmt.Errorf("usage: loomi runs attach <run-id>")
		}
		return fmt.Errorf("usage: loomi runs follow <run-id>")
	}
	if *output != "text" && *output != "json" {
		return fmt.Errorf("--output must be text or json")
	}
	client := cli.NewClient(*host)
	runID := fs.Arg(0)
	renderer := cli.Renderer{Out: stdout}
	resumeAfter := *after
	if replay {
		run, err := client.GetRun(ctx, runID)
		if err != nil {
			return err
		}
		if *output == "json" {
			if err := renderer.PrintJSONLine(run); err != nil {
				return err
			}
		} else if err := renderer.PrintRun(run); err != nil {
			return err
		}
		events, err := client.ListEvents(ctx, runID, resumeAfter)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.Sequence > resumeAfter {
				resumeAfter = event.Sequence
			}
			if err := printStreamedEvent(renderer, event, *output, *compact, *toolsOnly); err != nil {
				return err
			}
		}
		if isCLIRunTerminal(run.Status) {
			return nil
		}
	} else if resumeAfter < 0 {
		events, err := client.ListEvents(ctx, runID, 0)
		if err != nil {
			return err
		}
		resumeAfter = maxEventSequence(events)
	}
	return client.StreamEvents(ctx, runID, resumeAfter, func(event cli.RunEvent) {
		_ = printStreamedEvent(renderer, event, *output, *compact, *toolsOnly)
	})
}

func isCLIRunTerminal(status string) bool {
	switch status {
	case "completed", "failed", "stopped", "cancelled":
		return true
	default:
		return false
	}
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

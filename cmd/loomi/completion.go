package main

import (
	"fmt"
	"io"
	"strings"
)

var completionCommands = []string{
	"help",
	"version",
	"completion",
	"doctor",
	"status",
	"config",
	"chat",
	"sessions",
	"models",
	"personas",
	"tools",
	"mcp",
	"lsp",
	"artifacts",
	"memory",
	"agent",
	"browser",
	"run",
	"runs",
	"events",
	"approvals",
}

func cmdCompletion(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: loomi completion <bash|zsh|fish>")
	}
	switch args[0] {
	case "bash":
		return printBashCompletion(stdout)
	case "zsh":
		return printZshCompletion(stdout)
	case "fish":
		return printFishCompletion(stdout)
	default:
		return fmt.Errorf("unsupported shell %s", args[0])
	}
}

func printBashCompletion(stdout io.Writer) error {
	commands := strings.Join(completionCommands, " ")
	_, err := fmt.Fprintf(stdout, `_loomi_completion() {
  local cur prev commands
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="%s"

  if [[ ${COMP_CWORD} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
    return 0
  fi

  case "${COMP_WORDS[1]}" in
    help)
      COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh fish" -- "${cur}") )
      ;;
    config)
      COMPREPLY=( $(compgen -W "show set unset" -- "${cur}") )
      ;;
    sessions)
      COMPREPLY=( $(compgen -W "list resume" -- "${cur}") )
      ;;
    models|personas|tools)
      COMPREPLY=( $(compgen -W "list" -- "${cur}") )
      ;;
    lsp)
      COMPREPLY=( $(compgen -W "tools" -- "${cur}") )
      ;;
    artifacts)
      COMPREPLY=( $(compgen -W "list read" -- "${cur}") )
      ;;
    memory)
      COMPREPLY=( $(compgen -W "list search show audit" -- "${cur}") )
      ;;
    agent)
      COMPREPLY=( $(compgen -W "tasks tools" -- "${cur}") )
      ;;
    browser)
      COMPREPLY=( $(compgen -W "tools events" -- "${cur}") )
      ;;
    mcp)
      COMPREPLY=( $(compgen -W "servers" -- "${cur}") )
      ;;
    runs)
      COMPREPLY=( $(compgen -W "status stop attach follow" -- "${cur}") )
      ;;
    events)
      COMPREPLY=( $(compgen -W "tail" -- "${cur}") )
      ;;
    approvals)
      COMPREPLY=( $(compgen -W "list follow approve deny" -- "${cur}") )
      ;;
  esac
}
complete -F _loomi_completion loomi
`, commands, commands)
	return err
}

func printZshCompletion(stdout io.Writer) error {
	_, err := fmt.Fprint(stdout, `#compdef loomi

_loomi() {
  local -a commands
  commands=(
    'help:show command help'
    'version:show CLI version'
    'completion:print shell completion script'
    'doctor:check API, config, provider, and tools'
    'status:check local Loomi API'
    'config:show or write local CLI defaults'
    'chat:open an interactive Loomi chat shell'
    'sessions:list or resume threads'
    'models:list model providers'
    'personas:list personas'
    'tools:list runtime tools'
    'mcp:list safe MCP server status'
    'lsp:list LSP tool catalog entries'
    'artifacts:list or read thread artifacts'
    'memory:list, search, show, or audit safe memory'
    'agent:list coordination-only tasks and tools'
    'browser:list browser tools and run events'
    'run:start a run and stream events'
    'runs:show or stop runs'
    'events:stream run events'
    'approvals:list, follow, approve, or deny tool approvals'
  )

  _arguments \
    '1:command:->command' \
    '*::arg:->arg'

  case $state in
    command)
      _describe 'command' commands
      ;;
    arg)
      case $words[2] in
        help) _values 'topic' help version completion doctor run tools mcp lsp artifacts memory agent browser events runs approvals config ;;
        completion) _values 'shell' bash zsh fish ;;
        config) _values 'command' show set unset ;;
        sessions) _values 'command' list resume ;;
        models|personas|tools) _values 'command' list ;;
        mcp) _values 'command' servers ;;
        lsp) _values 'command' tools ;;
        artifacts) _values 'command' list read ;;
        memory) _values 'command' list search show audit ;;
        agent) _values 'command' tasks tools ;;
        browser) _values 'command' tools events ;;
        runs) _values 'command' status stop attach follow ;;
        events) _values 'command' tail ;;
        approvals) _values 'command' list follow approve deny ;;
      esac
      ;;
  esac
}

_loomi
`)
	return err
}

func printFishCompletion(stdout io.Writer) error {
	for _, line := range []string{
		"complete -c loomi -f",
		"complete -c loomi -n '__fish_use_subcommand' -a 'help' -d 'Show command help'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'version' -d 'Show CLI version'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'completion' -d 'Print shell completion script'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'doctor' -d 'Check API, config, provider, and tools'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'status' -d 'Check local Loomi API'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'config' -d 'Show or write local CLI defaults'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'chat' -d 'Open interactive chat shell'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'sessions' -d 'List or resume threads'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'models' -d 'List model providers'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'personas' -d 'List personas'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'tools' -d 'List runtime tools'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'mcp' -d 'List MCP server status'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'lsp' -d 'List LSP tools'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'artifacts' -d 'List or read thread artifacts'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'memory' -d 'List, search, show, or audit safe memory'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'agent' -d 'List coordination-only tasks and tools'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'browser' -d 'List browser tools and run events'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'run' -d 'Start a run and stream events'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'runs' -d 'Show or stop runs'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'events' -d 'Stream run events'",
		"complete -c loomi -n '__fish_use_subcommand' -a 'approvals' -d 'Manage tool approvals'",
		"complete -c loomi -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'",
		"complete -c loomi -n '__fish_seen_subcommand_from config' -a 'show set unset'",
		"complete -c loomi -n '__fish_seen_subcommand_from sessions' -a 'list resume'",
		"complete -c loomi -n '__fish_seen_subcommand_from models personas tools' -a 'list'",
		"complete -c loomi -n '__fish_seen_subcommand_from mcp' -a 'servers'",
		"complete -c loomi -n '__fish_seen_subcommand_from lsp' -a 'tools'",
		"complete -c loomi -n '__fish_seen_subcommand_from artifacts' -a 'list read'",
		"complete -c loomi -n '__fish_seen_subcommand_from memory' -a 'list search show audit'",
		"complete -c loomi -n '__fish_seen_subcommand_from agent' -a 'tasks tools'",
		"complete -c loomi -n '__fish_seen_subcommand_from browser' -a 'tools events'",
		"complete -c loomi -n '__fish_seen_subcommand_from runs' -a 'status stop attach follow'",
		"complete -c loomi -n '__fish_seen_subcommand_from events' -a 'tail'",
		"complete -c loomi -n '__fish_seen_subcommand_from approvals' -a 'list follow approve deny'",
	} {
		if _, err := fmt.Fprintln(stdout, line); err != nil {
			return err
		}
	}
	return nil
}

// Package ai streams quick answers from Claude for the launcher's AI mode.
//
// Two backends are supported, chosen automatically:
//   - "api": the Anthropic SDK, used when ANTHROPIC_API_KEY or
//     ANTHROPIC_AUTH_TOKEN is set.
//   - "claude": the Claude Code CLI in print mode, which reuses the user's
//     existing login and needs no key.
//
// OMNIA_AI_BACKEND forces a backend and OMNIA_AI_MODEL overrides the model.
package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	DefaultModel = "claude-opus-5"
	maxTokens    = 4096

	systemPrompt = "You are the quick-answer assistant inside a keyboard launcher. " +
		"The user wants a fast, direct answer. Reply in plain text without markdown headings, " +
		"tables or code fences unless code is explicitly requested. Lead with the answer, " +
		"keep it short, and skip preamble."
)

// Chunk is one streamed piece of an answer. Done is set on the final chunk;
// Err is set when the stream ended in failure.
type Chunk struct {
	Text string
	Done bool
	Err  error
}

// Backend identifies how answers are produced.
type Backend string

const (
	BackendNone   Backend = ""
	BackendAPI    Backend = "api"
	BackendClaude Backend = "claude"
)

// Detect picks a backend from the environment.
func Detect() Backend {
	switch strings.ToLower(os.Getenv("OMNIA_AI_BACKEND")) {
	case "api":
		return BackendAPI
	case "claude", "cli":
		return BackendClaude
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		return BackendAPI
	}
	if _, err := exec.LookPath("claude"); err == nil {
		return BackendClaude
	}
	return BackendNone
}

// Model returns the model to use.
func Model() string {
	if m := os.Getenv("OMNIA_AI_MODEL"); m != "" {
		return m
	}
	return DefaultModel
}

// Ask streams an answer for prompt. The returned channel is closed after the
// final chunk (which has Done set). Cancel ctx to abort.
func Ask(ctx context.Context, prompt string) <-chan Chunk {
	ch := make(chan Chunk, 32)
	go func() {
		defer close(ch)
		var err error
		switch Detect() {
		case BackendAPI:
			err = askAPI(ctx, prompt, ch)
		case BackendClaude:
			err = askClaudeCLI(ctx, prompt, ch)
		default:
			err = errors.New("no AI backend: set ANTHROPIC_API_KEY or install the claude CLI")
		}
		if ctx.Err() != nil {
			return
		}
		ch <- Chunk{Done: true, Err: err}
	}()
	return ch
}

func askAPI(ctx context.Context, prompt string, ch chan<- Chunk) error {
	client := anthropic.NewClient()
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(Model()),
		MaxTokens: maxTokens,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	for stream.Next() {
		event := stream.Current()
		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if text, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok {
				select {
				case ch <- Chunk{Text: text.Text}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	return stream.Err()
}

func askClaudeCLI(ctx context.Context, prompt string, ch chan<- Chunk) error {
	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
		"--no-session-persistence",
		"--append-system-prompt", systemPrompt,
	}
	if m := os.Getenv("OMNIA_AI_MODEL"); m != "" {
		args = append(args, "--model", m)
	}
	cmd := exec.CommandContext(ctx, "claude", args...)
	// Run from $HOME so the CLI does not pick up project instructions from the cwd.
	if home, err := os.UserHomeDir(); err == nil {
		cmd.Dir = home
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var (
		streamed bool
		final    string
		cliErr   string
	)
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev cliEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "stream_event":
			if ev.Event.Type == "content_block_delta" && ev.Event.Delta.Type == "text_delta" {
				streamed = true
				select {
				case ch <- Chunk{Text: ev.Event.Delta.Text}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		case "result":
			if ev.IsError {
				cliErr = ev.Result
			} else if !streamed {
				final = ev.Result
			}
		}
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if cliErr != "" {
		return errors.New(strings.TrimSpace(cliErr))
	}
	if waitErr != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("claude: %s", firstLine(msg))
	}
	if !streamed && final != "" {
		ch <- Chunk{Text: final}
	}
	return nil
}

type cliEvent struct {
	Type    string `json:"type"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
	Event   struct {
		Type  string `json:"type"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	} `json:"event"`
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

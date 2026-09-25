package ai

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	maxPromptRunes    = 1000
	maxSelectionRunes = 16000
)

var (
	ErrNotConfigured    = errors.New("ai is not configured")
	ErrInvalidAction    = errors.New("invalid ai action")
	ErrPromptRequired   = errors.New("prompt is required")
	ErrSelectionMissing = errors.New("selection is required")
	ErrPromptTooLong    = errors.New("prompt is too long")
	ErrSelectionTooLong = errors.New("selection is too long")
)

type Config struct {
	APIKey          string
	Model           string
	DailyTokenLimit int
}

type Service struct {
	client *openRouter
}

func NewService(cfg Config) *Service {
	if cfg.APIKey == "" || cfg.Model == "" {
		return &Service{}
	}
	return &Service{client: newOpenRouter(cfg.APIKey, cfg.Model)}
}

const baseRules = "Reply with the result only: no preamble, closing remarks, explanations, or wrapping quotes. " +
	"Always reply in the same language as the user's text or instruction. " +
	"Notes are stored as HTML, so reply with an HTML fragment, not markdown. " +
	"Allowed tags: p, h1, h2, h3, ul, ol, li, blockquote, strong, em, s, code, a (with href), br. " +
	"For checklists or to-dos, use exactly this format: " +
	`<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>item text</p></li></ul> ` +
	`and set data-checked="true" only for items that are actually done. ` +
	"Do not invent other tags, do not use style or class attributes, and do not wrap the output in <html>, <body>, or a code fence."

const editRules = "The user's text is an HTML fragment. Preserve its structure and formatting (bold, links, lists, checklists) unless the instruction says otherwise. " +
	"If the user's text is not wrapped in a block tag (p, h1-h3, ul, ol, blockquote), reply without block tags too. "

var fixedActions = map[string]string{
	"rewrite":     "Rewrite the user's text to be clearer and easier to read, keeping the same meaning. " + editRules,
	"shorten":     "Shorten the user's text without losing its key points. " + editRules,
	"fix_grammar": "Fix the spelling and grammar of the user's text without changing its meaning or writing style. " + editRules,
}

func (s *Service) Build(in StreamInput) ([]message, error) {
	if s.client == nil {
		return nil, ErrNotConfigured
	}

	prompt := strings.TrimSpace(in.Prompt)
	selection := strings.TrimSpace(in.Selection)

	if utf8.RuneCountInString(prompt) > maxPromptRunes {
		return nil, ErrPromptTooLong
	}
	if utf8.RuneCountInString(selection) > maxSelectionRunes {
		return nil, ErrSelectionTooLong
	}

	if instruction, ok := fixedActions[in.Action]; ok {
		if selection == "" {
			return nil, ErrSelectionMissing
		}
		return []message{
			{Role: "system", Content: instruction + baseRules},
			{Role: "user", Content: selection},
		}, nil
	}

	if in.Action != "ask" {
		return nil, ErrInvalidAction
	}
	if prompt == "" {
		return nil, ErrPromptRequired
	}

	if selection == "" {
		return []message{
			{Role: "system", Content: "You are a writing assistant inside a notes app. Write what the user asks for, using a fitting structure: paragraphs, headings, lists, or a checklist when requested. " + baseRules},
			{Role: "user", Content: prompt},
		}, nil
	}

	return []message{
		{Role: "system", Content: "You are a writing assistant inside a notes app. The user gives you a text and an instruction. Apply the instruction to the text and return the final result. " + editRules + baseRules},
		{Role: "user", Content: "Text (HTML):\n" + selection + "\n\nInstruction:\n" + prompt},
	}, nil
}

func (s *Service) Stream(ctx context.Context, messages []message, onDelta func(string) error) (int, error) {
	return s.client.stream(ctx, messages, onDelta)
}

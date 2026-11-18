// mautrix-signal - A Matrix-Signal puppeting bridge.
// Copyright (C) 2024 Tulir Asokan
// Copyright (C) 2026 Dominik Sattler (preserve-deletions-and-edits fork)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package preserve

import (
	"testing"
	"time"
)

var testTemplates = Templates{
	Redacted:        "🗑️ Deleted by sender • {timestamp}",
	Edited:          "✏️ Edited • {timestamp}",
	EditedMulti:     "✏️ Edited ({count}×) • {timestamp}",
	OriginalLabel:   "Original:",
	VersionLabel:    "v{n}:",
	TimestampFormat: "2006-01-02 15:04",
	MaxEditHistory:  3,
}

// fixedTime: 2026-05-12 22:14:00 UTC — picked to match spec examples.
var fixedTime = time.Date(2026, 5, 12, 22, 14, 0, 0, time.UTC)

func TestWrapStrikethrough_SingleLine(t *testing.T) {
	got := WrapStrikethrough("Hallo, wie geht's?", "")
	if got.Body != "~~Hallo, wie geht's?~~" {
		t.Errorf("Body = %q, want %q", got.Body, "~~Hallo, wie geht's?~~")
	}
	if got.FormattedBody != "<del>Hallo, wie geht&#39;s?</del>" {
		t.Errorf("FormattedBody = %q, want HTML-escaped <del>...</del>", got.FormattedBody)
	}
}

func TestWrapStrikethrough_MultiLine(t *testing.T) {
	got := WrapStrikethrough("zeile1\nzeile2", "")
	want := "~~zeile1~~\n~~zeile2~~"
	if got.Body != want {
		t.Errorf("Body = %q, want %q", got.Body, want)
	}
}

func TestWrapStrikethrough_EmptyLinesPreserved(t *testing.T) {
	// Empty lines should not be wrapped in tildes (would render as ~~~~).
	got := WrapStrikethrough("a\n\nb", "")
	want := "~~a~~\n\n~~b~~"
	if got.Body != want {
		t.Errorf("Body = %q, want %q", got.Body, want)
	}
}

func TestWrapStrikethrough_FormattedPassthrough(t *testing.T) {
	// When formattedIn is provided, it is wrapped directly (no re-escaping of
	// markup the upstream converter already produced).
	got := WrapStrikethrough("hello", "<strong>hello</strong>")
	if got.FormattedBody != "<del><strong>hello</strong></del>" {
		t.Errorf("FormattedBody = %q, want <del><strong>hello</strong></del>", got.FormattedBody)
	}
}

func TestRenderTimestamp(t *testing.T) {
	got := RenderTimestamp(fixedTime, "2006-01-02 15:04")
	if got != "2026-05-12 22:14" {
		t.Errorf("RenderTimestamp = %q, want %q", got, "2026-05-12 22:14")
	}
	if RenderTimestamp(fixedTime, "") != "" {
		t.Errorf("RenderTimestamp with empty layout should return empty")
	}
}

func TestBuildRedactionMarker(t *testing.T) {
	got := BuildRedactionMarker(testTemplates, fixedTime)
	want := "🗑️ Deleted by sender • 2026-05-12 22:14"
	if got.Body != want {
		t.Errorf("Body = %q, want %q", got.Body, want)
	}
	// Emoji and bullet are not HTML-special and should not be escaped.
	if got.FormattedBody != want {
		t.Errorf("FormattedBody = %q, want %q", got.FormattedBody, want)
	}
}

func TestBuildEditMarker_Single(t *testing.T) {
	got := BuildEditMarker(testTemplates, fixedTime, 1)
	want := "✏️ Edited • 2026-05-12 22:14"
	if got.Body != want {
		t.Errorf("Body = %q, want %q", got.Body, want)
	}
}

func TestBuildEditMarker_Multi(t *testing.T) {
	got := BuildEditMarker(testTemplates, fixedTime, 3)
	want := "✏️ Edited (3×) • 2026-05-12 22:14"
	if got.Body != want {
		t.Errorf("Body = %q, want %q", got.Body, want)
	}
}

func TestBuildVersionLabel(t *testing.T) {
	if got := BuildVersionLabel(testTemplates, 2); got != "v2:" {
		t.Errorf("BuildVersionLabel(2) = %q, want %q", got, "v2:")
	}
}

func TestAppendBlock(t *testing.T) {
	base := Fragment{
		Body:          "Hallo",
		FormattedBody: "Hallo",
	}
	marker := Fragment{
		Body:          "✏️ Edited",
		FormattedBody: "✏️ Edited",
	}
	got := base.AppendBlock(marker)
	if got.Body != "Hallo\n✏️ Edited" {
		t.Errorf("Body = %q, want %q", got.Body, "Hallo\n✏️ Edited")
	}
	if got.FormattedBody != "Hallo<br>✏️ Edited" {
		t.Errorf("FormattedBody = %q, want %q", got.FormattedBody, "Hallo<br>✏️ Edited")
	}
}

func TestAppendBlock_EmptyReceiver(t *testing.T) {
	// Appending to an empty fragment should produce just the block (no leading
	// separator) — matters when building a marker-only annotation.
	empty := Fragment{}
	block := Fragment{Body: "marker", FormattedBody: "marker"}
	got := empty.AppendBlock(block)
	if got.Body != "marker" {
		t.Errorf("Body = %q, want %q", got.Body, "marker")
	}
	if got.FormattedBody != "marker" {
		t.Errorf("FormattedBody = %q, want %q", got.FormattedBody, "marker")
	}
}

// TestSpecExample reproduces the multi-edit example from the spec end-to-end
// with the helper primitives, as a sanity check that the building blocks
// compose into the documented target shape.
func TestSpecExample(t *testing.T) {
	v2Time := time.Date(2026, 5, 12, 22, 14, 0, 0, time.UTC)
	v1Time := time.Date(2026, 5, 12, 22, 10, 0, 0, time.UTC)
	now := time.Date(2026, 5, 12, 22, 18, 0, 0, time.UTC)

	current := Fragment{
		Body:          "Hallo, wie läufts denn so?",
		FormattedBody: "Hallo, wie läufts denn so?",
	}
	marker := BuildEditMarker(testTemplates, now, 2)

	// v2: previous version, plain
	v2Line := Fragment{
		Body:          BuildVersionLabel(testTemplates, 2) + " Hallo, wie läufts? (" + RenderTimestamp(v2Time, "15:04") + ")",
		FormattedBody: BuildVersionLabel(testTemplates, 2) + " Hallo, wie läufts? (" + RenderTimestamp(v2Time, "15:04") + ")",
	}
	// v1: original, strikethrough'd
	v1Strike := WrapStrikethrough("Hallo, wie geht's?", "")
	v1Line := Fragment{
		Body:          BuildVersionLabel(testTemplates, 1) + " " + v1Strike.Body + " (" + RenderTimestamp(v1Time, "15:04") + ")",
		FormattedBody: BuildVersionLabel(testTemplates, 1) + " " + v1Strike.FormattedBody + " (" + RenderTimestamp(v1Time, "15:04") + ")",
	}

	result := current.AppendBlock(marker).AppendBlock(v2Line).AppendBlock(v1Line)

	wantBody := "Hallo, wie läufts denn so?\n" +
		"✏️ Edited (2×) • 2026-05-12 22:18\n" +
		"v2: Hallo, wie läufts? (22:14)\n" +
		"v1: ~~Hallo, wie geht's?~~ (22:10)"
	if result.Body != wantBody {
		t.Errorf("composed Body =\n%s\nwant:\n%s", result.Body, wantBody)
	}
}

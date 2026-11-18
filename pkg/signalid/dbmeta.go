// mautrix-signal - A Matrix-Signal puppeting bridge.
// Copyright (C) 2024 Tulir Asokan
// Copyright (C) 2026 Dominik Sattler (preserve-deletions-and-edits fork)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package signalid

import (
	"go.mau.fi/util/jsontime"
)

type PortalMetadata struct {
	Revision               uint32 `json:"revision,omitempty"`
	ExpirationTimerVersion uint32 `json:"expiration_timer_version,omitempty"`
	// Lazy resync tracking
	LastSync jsontime.Unix `json:"last_sync,omitempty"`
}

type MessageMetadata struct {
	ContainsAttachments bool     `json:"contains_attachments,omitempty"`
	MatrixPollOptionIDs []string `json:"matrix_poll_option_ids,omitempty"`

	// PreservedEdits stores the body of each previous revision of this message
	// when the fork-local preserve-deletions-and-edits feature is enabled. It
	// is nil/empty for messages received before the feature was switched on,
	// in which case annotations fall back to a "(original not preserved)"
	// placeholder. The `omitempty` tag keeps old DB rows binary-compatible.
	PreservedEdits []PreservedEdit `json:"preserved_edits,omitempty"`
}

// PreservedEdit captures one historical revision of a message — either an
// edit's previous body, or the body that existed before a sender-initiated
// delete-for-everyone. Used to render inline edit/redaction annotations.
type PreservedEdit struct {
	Body          string             `json:"body"`
	FormattedBody string             `json:"formatted_body,omitempty"`
	Timestamp     jsontime.UnixMilli `json:"timestamp"`
	// Deleted distinguishes a delete-for-everyone event from a regular edit
	// so the consumer can pick the redaction marker template instead of the
	// edit marker template.
	Deleted bool `json:"deleted,omitempty"`
}

type UserLoginMetadata struct {
	ChatsSynced     bool               `json:"chats_synced,omitempty"`
	LastContactSync jsontime.UnixMilli `json:"last_contact_sync,omitempty"`
}

type GhostMetadata struct {
	ProfileFetchedAt jsontime.UnixMilli `json:"profile_fetched_at"`
}

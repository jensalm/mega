---
title: hotkey for "allow"
state: researching
created: 2026-11-13
tags: [enhancement]
---

it would be really helpful if there were a hotkey for "allow" dialogs.

i'm continually hitting `ctrl+enter` to launch prompts. it would be helpful if i could hit `ctrl+enter` to approve dialogs, too. i don't know if it's possible but `ctrl+shift+enter` would be useful as a "allow for this session".

## background

Noted by the planning agent on 2026-11-13, for when this card is worked. Pending requests dock above the reply box, oldest first, and an approval's card has Allow, Deny and, where the harness offers it, "Allow this tool for the session". Things to settle: which request the key answers when several are docked (the oldest is the obvious one); that it answers only while the reply box is empty, so Ctrl+Enter with a draft still sends; that it never answers a question, only an approval; and that Ctrl+Shift+Enter does nothing where the harness offers no session-wide allow. The key is the operator's own answer, so `permissions-never-weakened` is untouched, but a key that approves must not be reachable by accident: no auto-repeat, and nothing on a request that appeared under the keypress.

One more thing to settle, added 2026-11-16: a pi dialog is always a question, never an approval, so the key never answers one.

Confirmed by Michael on 2026-11-16: Ctrl+Shift+Enter reaches the page in the Firefox-based browsers he uses, LibreWolf among them. Both keys work there.

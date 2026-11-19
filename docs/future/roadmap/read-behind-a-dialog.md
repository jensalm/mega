---
title: read behind a dialog
state: evaluating
created: 2026-11-18
tags: [defect]
---

The page acknowledges a conversation's output (`read`) while a dialog covers it, so it can leave needs-you before the operator has seen it.

## background

Found by terminus on 2026-11-18, while reviewing the reply box batch (`4090203bcdb4`, then `b129a9957a61`). The read effect in `ui/src/App.tsx` posts once the history is loaded and current and the tab is visible. It does not check whether a modal dialog covers the conversation.

- **Ctrl+/ behind a dialog** was the first path found. Michael's call that day: Ctrl+/ does nothing while a dialog is open, and that is built.
- **What remains:**
  - a switch by the address (the hash, the browser's back and forward) behind an open dialog;
  - new output in the conversation already open, arriving while a dialog such as the version panel is open.

The review's suggestion is to defer the read while a modal is open and post it when the last one closes, with a panel check that an unread target stays unread until the dialog closes.

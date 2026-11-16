---
title: conversation timestamps
state: evaluating
created: 2026-11-14
tags: [enhancement]
---

it would be really useful if the conversation were timestamped... could probably go right in the same hover-enabled spot where the model/run detail goes.

## background

Noted by the planning agent on 2026-11-14. Every event in a conversation's log already carries its time, so this is display only. The turn's line, shown on hover at the right end of the turn's separator, is the spot Michael means; it now reads "Ran model · effort · mode". Things to settle: whether the time is when the prompt was sent or when the turn ended, or the start and how long it took; how a time from another day reads; and whether the operator's own prompt block wants its time as well, since that is the row he looks for when coming back. One thing to check before choosing how a time is shown. Michael's browser is LibreWolf, whose hardened defaults may report the time zone as UTC to the page. A time formatted by the browser would then be hours off from his clock. If his LibreWolf does that, the service should say what zone it runs in and the page should format with that, not with the browser's own.

Built with `hotkey-for-allow` and `action-view-improvements`, which change the same surface.

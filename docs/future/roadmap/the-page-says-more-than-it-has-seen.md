---
title: the page says more than it has seen
state: building
created: 2026-11-18
tags: [defect]
---

Five places where mega claims something it has not observed: output marked read that was never on screen, a page that says it is live before it has caught up, and a release that reports a clean stop it could not check. Each is small. Together they are what stands between the reply box work and a clean review.

## why

Terminus reviewed the reply box batch and the fix for reads behind a dialog on 2026-11-18 and did not come back clean. It confirmed both pieces of work and blocked on earlier code in the files they touch. None of the five was caused by either batch. Michael's direction: deal with these before returning to pull request 3.

## the five

1. **Output read below the view.** With the history scrolled up and "Newer activity below" showing, new output is still marked read, though it is not on screen. The same happens when a dialog closes over a scrolled-up view. The control that says there is something unseen and the mark that says there is nothing unread disagree. A read should wait until the view is back at the newest output.
2. **A failed read is not retried** until something else changes, so output the operator has seen can stay unread after one lost request.
3. **Requests stay answerable while the page is stale,** after a reconnect and before the workspace has been read again. The service refuses an answer to a request that is no longer pending, so nothing wrong is delivered, but the page offers a button it cannot stand behind, where its other controls are already held while stale.
4. **The footer and the rail say "Live"** after a reconnect, before the workspace has been read again. For that moment the page is showing what it knew before.
5. **A release reports a clean stop when it could not look.** Codex and pi find leftover child processes by reading `/proc`. Where that cannot be read, as on macOS, the scan is taken to mean nothing survived. It should say the check could not be made. This is the honest half of the process gap on `mega-on-macos-and-windows`, and it needs no Mac: a real listing there can wait for someone who has one.

Also, an advisory: the first-run suite's check of the start-up log leaves pi out, though its steps say each harness.

## what does not change

Opening a conversation still moves the reading position to what is on screen. Reading never answers a request. A tab that is not visible reads nothing, and neither does a page with a dialog open.

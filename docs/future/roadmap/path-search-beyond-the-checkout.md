---
title: path search beyond the checkout
state: researching
created: 2026-11-14
tags: [enhancement, spike]
---

The reply box's `@` search can also reach directories outside the conversation's checkout that the operator has named, such as `~/Sandbox`, typed and shown in `~` notation. Without any named, it behaves exactly as now.

## why

Michael, 2026-11-14, on `path-search-in-the-input-panel`: "i would like to make this less restrictive. ideally it would allow me to reference paths in `~/Sandbox`, using `~` notation. maybe we could include optional external directories like that?" His habit is to drop a screenshot or a scratch file in `~/Sandbox` or `~/Screenshots` and point an agent at it. The search stops at the checkout by design, so those paths are typed by hand, and they are the ones with names nobody remembers.

## shape

Accepted by Michael on 2026-11-14 as proposed; the spike settles what is inserted.


- **Named by the operator, off by default.** A list of extra directories in mega's config file, for example `~/Sandbox` and `~/Screenshots`. They are the operator's own places, not a project's, so they apply in every project. Nothing outside the checkout is ever listed unless it is under one of them. This is not a file browser for the whole machine.
- **Reached by how the path starts.** `@` followed by `~/` or `/` searches the named directories; anything else searches the checkout as now. So ordinary results stay free of screenshots, and `@~/Sa` completes to the named directories that match.
- **Newest first there.** In the checkout, the best match by name wins. In a sandbox the file wanted is nearly always the one just made, so within a named directory results are ordered by modification time, newest first, then filtered by what is typed.
- **Shown with `~`.** The list shows `~/Sandbox/truncated.png`. What is inserted is settled by the spike.
- **The same limits as the checkout listing.** Names only, never contents. A bounded walk that skips dot-directories. No following a symlink out of the named directory. The same cap on results.

## spike

Recorded in the capability report with versions, as the `@path` follow-up of 2026-11-10 was:

- Whether Claude attaches a file named as `@~/dir/file` or `@/abs/dir/file` in a programmatic prompt, as it does a relative `@path`, and which form survives a path with spaces.
- What a Codex agent does with a bare `~/dir/file` against a bare absolute path.
- What each harness asks before reading outside the checkout. Mega changes nothing there: an approval the harness would ask for in a terminal is still asked.

The inserted text follows from that: the form each harness reliably resolves, which may be the absolute path even though the list shows `~`.

## not this card

Per-project extra directories, which can follow if the global list is not enough, and attaching a file from the browser, by drop or paste, which is a different mechanism with the same purpose and deserves its own card if Michael wants it.

## checks

A service test for the listing: a named directory, a path outside any named directory refused, dot-directories skipped, a symlink out not followed, newest first. In the panel suite: `@~/` opening the list, picking, and the inserted text per harness.

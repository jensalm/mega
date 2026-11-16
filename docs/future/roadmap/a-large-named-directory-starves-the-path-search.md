---
title: a large named directory starves the path search
state: building
created: 2026-11-16
tags: [defect]
---

With two directories named for the reply box's `@~/` search, typing `@~/Screenshots/` shows the folder itself and none of the files in it. The search should list what is in the folder the operator has typed, however large any other named directory is.

## why

Found by Michael on 2026-11-16, minutes after `path_search.directories` was first set on his machine to `~/Sandbox` and `~/Screenshots`. `~/Screenshots` holds four files. `~/Sandbox` holds about 76,000 entries, most of them in one deep tree. Asked for `~/Screenshots/`, the running service answered with four results: the `~/Screenshots/` folder, and three files deep inside `~/Sandbox` whose paths happen to match the typed letters.

Two things combine:

- **One bound covers every named directory.** The walk is bounded, as the card asked, and the directories are walked in the order the config names them. The first one used the whole bound, so the second was listed only as itself.
- **What is typed is only a loose filter.** The implementer made the typed text a filter over the listing and never a path, so that nothing outside a named directory could be reached. That is safe, and it means typing a folder's own path does not look inside that folder.

The same bound hurts the large directory too. A walk that stops part of the way through cannot promise that the newest file is among what it saw, and the newest file is the one this feature is for.

The planning agent's check of the setting, before Michael restarted, counted four results for `~/Screenshots/` and reported them as that folder's files without looking at where they came from. They were these four.

## shape

- **Typed text that names a place inside a named directory is looked in directly.** `@~/Screenshots/` lists that folder; `@~/Sandbox/shots/` lists that one. The text is resolved, links and all, and is used only if the result is still inside a named directory. Anything else falls back to the filter, as now. Nothing outside a named directory is listed either way.
- **Shallow first.** Within a directory, its own files come before anything deeper, newest first, so a file just dropped at the top is always there whatever lies beneath.
- **Each named directory has its own bound,** so one cannot use up another's.
- **The list says when it stopped short,** so a missing file is not taken to be absent.

## checks

A service test with a small directory named after one holding more entries than the bound: the small one's files are listed; a file newly added at the top of the large one is listed; a typed path that leaves the named directory through a link or `..` lists nothing. In the panel suite: `@~/second/` showing that folder's files.

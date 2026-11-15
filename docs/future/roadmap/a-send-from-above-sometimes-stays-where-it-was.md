---
title: a send from above sometimes stays where it was
state: building
created: 2026-11-14
tags: [defect]
---

Sending a prompt while scrolled up should take the view to the end and follow the reply. About one time in four it does not: the view goes to the end as the box closes, then stops following, and the prompt and its reply arrive below the fold.

## why

Found on 2026-11-14 while reviewing the first-run batch. The panel suite's step "the view went to the end and followed the reply" scrolls to the top, sends with Ctrl+Enter and waits for the reply. It failed 5 runs in 12 on the working tree and 2 in 12 on the committed tree at fd880aa, built apart in scratch, each time "470px from the bottom". So it is not from the first-run batch. The implementer's run of every suite passed, as a run usually does.

The constant distance says the view was pinned once, before the prompt row arrived, and then nothing followed. That is `following` in `ui/src/latest.ts` going false straight after `toLatest` set it. Sending from the bottom is not affected in the same way, since nothing has to move first, which is why daily use has not shown it much.

## likely cause, not confirmed

`following` is set by every scroll event to whether the view is at the bottom at the moment the event is read. Scroll events arrive a frame after the scroll that caused them. `fit()` moves the view twice on a send from above: once to keep the reader's place as the box closes, then to the end. If content grows by more than the 32 pixel slack before a queued event from either is read, and nothing pinned the view in between, that event reads "not at the bottom" and switches following off for good. The dock fix of 2026-11-12 met the same trap once already, in `fit()`'s own measuring.

## shape

- Make leaving the bottom something only the reader does. The view stops following when it moves up by the reader's own scroll, a wheel, a key, a drag of the bar, and not when the content under it grew. One way: a scroll event that finds the view no higher than it was last left does not end following.
- Find which of the three batches since 2026-11-12 brought it in, the dock fix, the turn line or neither, by running the step repeatedly on each commit. It may be older and only now common enough to see.
- The step exists and is right. The fix is done when it passes fifty runs in a row, and a second step sends from above with the fake holding its reply, as the dock steps do.

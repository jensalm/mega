---
title: improve the adopt-a-session dialog
state: evaluating
created: 2026-11-18
tags: [enhancement]
---

we need to make the "adopt a session" dialog larger and give it a round of design improvements.

## background

Noted by the planning agent on 2026-11-18. The dialog today is 480 pixels wide, as every dialog is. It has a select for the harness, then one radio row per session with the first prompt, a date and the session's id, then a name field and a warning about open terminals. A checkout with many sessions is a long list of look-alike rows in a narrow column, and the only clue to which session is which is how it began.

**A shape to react to:**

- **Larger.** Wide enough for two columns and most of the window's height, with the list scrolling inside and the heading and the Adopt button staying put.
- **All harnesses at once.** One list, each row carrying its harness mark, or a tab per harness with a count. Not a select that hides two thirds of what is there. A harness that cannot run here is said so.
- **Rows that tell sessions apart.** Newest first. How long ago, in words. The first prompt on two lines. A filter box.
- **A look before adopting.** Picking a row shows, beside the list, how the session began and how it left off: the last prompt and the end of the last reply. That is the question the dialog is really asked: which of these is the one I mean.
- **What is not shown, said.** Sessions a conversation already holds are left out today with no word; a count of them would say why a session someone expects is missing.
- **The warnings where they apply.** The terminal warning for all, and for pi the plain statement that it has no lock.

**To settle.** A session today is known by its id, its first prompt and two times. Showing how it left off, a turn count or the model means each adapter reading more of the native session. That is a read of a private transcript, shown only to the operator, and it costs time on a checkout with many sessions; it should be read when a row is picked, not for every row.

## settled with Michael, 2026-11-18

- **A defect to fix first:** "the radio buttons seem to wrap strangely on some screens." The rows have to lay out properly at every width the page is used at.
- **Per harness stays.** One harness's sessions at a time is fine; the control that chooses the harness should show how many each has.
- **The look before adopting is wanted,** read when a row is picked.

---
title: focus on operator input
state: evaluating
created: 2026-11-18
tags: [enhancement]
---

when switching to a new agent, the focus should automatically be placed on the operator input. should we include a keyboard shortcut to focus on the operator input?

## background

Noted by the planning agent on 2026-11-18. Two things about the reply box bear on this:

- **Focus opens it.** The box rests at one line and opens to its full height when focus enters, and the history above moves to make room. Focusing it on every switch means every conversation opens with the box open.
- **Focus takes the keys.** With the cursor in the box, the arrow keys, Page Up and the space bar go to the box, not to the history. Someone who switches to a conversation to read it, not to write, would have to leave the box first.

**A shape to react to:**

- Switching to a conversation puts the focus in the reply box, as asked, when the switch was made by keyboard or by picking the conversation in the side panel.
- The box does not open until something is typed, so a switch does not move the history.
- A shortcut puts the focus in the reply box from anywhere on the page, and Escape leaves it again for the history. The two shortcuts the page has now are Ctrl+/ and Ctrl+'; a neighbour of those keeps them together.
- The Allow keys are unchanged: they already work from the reply box while it is empty.

## settled with Michael, 2026-11-18

- **The shortcut is Ctrl+Space.** It puts the focus in the reply box from anywhere on the page. Some desktops keep that key for switching input methods, so it is his to confirm in his own browser, as the other shortcuts were.
- **The box stays closed on focus until typing.** Focus alone no longer opens it. It opens at the first character, and it is open whenever it holds a draft.

---
title: agent hand-off turns
state: evaluating
created: 2026-11-14
tags: [enhancement]
---

i'll often bounce between a planning agent and an implementation agent... it would be useful if there was some kind of an indicator of which agent in a project was the last agent i interacted with.

or... other approaches that help me to come back to a project and know which agent i left off with.

## background

Noted by the planning agent on 2026-11-14. The small version is cheap and needs nothing new from the service: each conversation's log already says when Michael last sent it a prompt, so the side panel can mark, in each project, the conversation he last spoke to, and say how long ago. That is a fact about him, not attention, so it takes no mark from the attention set and never raises anything (`bookkeeping-never-creates-attention`).

The larger question in the card's second line, how to come back to a project and know where things stand, is what `re-entry-briefing` and `agents-roles-and-lifecycles-in-a-project` are about. This card stays the small indicator; what it teaches goes to those.

Not the same as `hand-off-between-conversations`, which is about carrying a return from one agent to another.

## shape, settled with Michael on 2026-11-17

The minimal version: keep track of the agent that was last instructed to do something, and show it.

- **What counts.** A prompt of any kind, sent through mega: a launch's opening prompt, and any prompt after it. Nothing else. Answering a request, starting a runtime, switching a mode and opening the conversation do not count. Michael: "what is helpful is keeping track of the last agent that was instructed to do something."
- **One per project.** In each project, the conversation most recently instructed carries the marker. That is enough until a richer structure ties agents into teams within a project, which is `agents-roles-and-lifecycles-in-a-project`.
- **What it looks like.** A small muted glyph after the conversation's name in the side panel. A running turn's elapsed time stays at the right of the row, after it. Hovering gives the time.
- **Kept by the service,** so it reads the same from every machine and through a share.
- **Not attention.** It uses no mark from the attention set, raises nothing, counts nothing and reorders nothing. What an agent does never moves it.

An assumption, to be confirmed: a prompt imported from outside mega, after a terminal visit, does not count, because an imported turn's time is when mega imported it and not when it was sent. That can change when `imported-turns-carry-their-own-times` lands.

---
title: change digestion
state: horizon
created: 2026-11-18
tags: [epic, spike]
---

Mega presents a change as a story: what was asked, what was decided along the way, what was done and why, and what was checked, with the code as the evidence and not the whole account. The story leads with the parts of the change that matter most and gives the noisy parts less weight. It hides nothing: the raw diff is always there to check.

## why

From Michael's presentation of mega, 2026-11-18, the fourth of seven future directions: "first class change digestion model in mega. today we 'digest' change by looking at diffs and reading code with only the input design as context. mega will gain first-class change context tooling... changes presented as a 'story'."

The agents now produce change faster than one person can read it. A diff says what changed and nothing about why, what was tried first, which finding forced the third rewrite, or what the reviewer accepted. All of that exists. It is in the conversation that made the change, the card it was built to, the review rounds it went through and the journal entry written after. Today the operator assembles it in his head, or reads an implementer's report and trusts it.

Michael ties this to how far attention can spread: "as quality management and change digestion get better, attention scales wider."

## weighting comes first

Michael, the same day, on reading this card: "part of the work there is also finding the most important aspects of the change are and featuring those, while reducing the weight of the noisier parts of the change. with that weighting in place, the change will be presented as a coherent story that explains what was changed and why."

So the story is the result, and the weighting is the work. A change is rarely even. A few lines carry the decision: a rule that changed, an interface that moved, a guard that was added. Around them sit the lines that follow from it: renames, call sites brought in step, tests adjusted, documents updated, generated files. A diff gives every line the same weight and orders them by file name. A reader has to find the few among the many each time.

Digesting a change means doing that finding once and showing the result:

- **Featured:** the parts that carry the change's intent, each with why it is there.
- **Lighter:** the parts that follow mechanically from those, grouped and summarised, such as "eleven call sites updated to the new signature".
- **Nothing dropped.** Every lighter group opens to the lines it stands for.

**What counts as important.** Michael, the same day: "i think 'important' is probably code changes... architectural changes being the most important." So there is an order, heaviest first:

1. **Architectural changes.** A new boundary or a moved one, a changed interface, a new dependency between parts, a rule about who owns what. One of these reshapes everything below it.
2. **Code changes.** Behaviour that is different after the change: new logic, a fixed path, a guard.
3. **What follows.** Call sites, tests, documents, configuration and generated files brought in step with the first two.

**How it is told.** His words: "the change should be presented in the way a developer would explain it to another developer, ideally." That is a walk-through, not a listing. A developer explaining a change starts with what they were trying to do, says which decision the rest hangs on, shows the piece of code that makes it real, and then says what else had to move because of it. They follow cause, not file names, and they say what they are unsure of.

**It must not hide anything.** His words: "this shouldn't hide any parts of the change... the operator can absolutely still check raw diffs." Weight decides what is read first, never what can be read. The raw diff is one step away from every part of the story, and the story says how much of the change it has summarised.

## what mega already holds

- **The conversation,** turn by turn, with every action and what it ran with.
- **The requests** the operator answered along the way, and how.
- **The implementer's report,** which is already a digest, written by the party with the most to gain from a kind reading.
- **Nothing of the review.** Terminus rounds scroll past as tool calls. The hub epic asks for a review to be a thing mega can show.
- **Nothing of the change itself.** Mega does not know which commits a conversation produced.

## what it has to settle

- **How an architectural change is recognised.** The order above is settled; telling the tiers apart is not. Some of it can be read from the change: a new exported name against a renamed local, an interface that changed against a body that did, hand-written against generated. Some of it only the record knows: what the card asked for, what the operator pushed back on, what a review finding forced. The second kind is why this belongs in mega and not in a diff viewer.
- **Whose voice the explanation is in.** The implementer could be asked to explain its own change this way, and already half does in its report. But it is the party with the most to gain from a kind reading, so its account has to be checked against the diff and the record before it is shown as the story.
- **What a wrong weighting costs.** A line that mattered and was filed under noise is the failure that counts. The lighter groups have to be cheap to open and honest about what is in them, and a part the weighting was unsure of should be featured, not buried.
- **What a story is made of,** and what ties its parts together: a card, a conversation or several, review rounds, commits.
- **Who writes it.** Assembled from records, it is trustworthy and dry. Written by a model, it reads well and needs checking against the records.
- **How it stays honest.** A story that leaves out the failed attempts is a press release. The record of what went wrong is the part worth most.
- **Where it sits beside review.** Terminus judges a change against the canon. This is the operator's own reading of it, made cheaper.

## depends on

Reviews as first-class things in mega, from `mega-as-the-development-hub`, and ranger's cards carrying more structure, which is the seventh direction and belongs to ranger.

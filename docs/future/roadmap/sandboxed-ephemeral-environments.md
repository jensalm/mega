---
title: sandboxed, ephemeral environments
state: horizon
created: 2026-11-18
tags: [epic, spike]
---

Agents run in isolated environments that are made for a piece of work and thrown away after it, reached over the overlay, with the same page and the same ease as an agent on the operator's own machine.

## why

From Michael's presentation of mega, 2026-11-18, the third of seven future directions: "sandboxed, ephemeral development environments. mega grows support for running agents over the overlay in isolated sandboxed environments, while keeping the same friction-free development experience."

Today every agent runs as the operator, on the operator's machine, in the operator's checkout. Each harness's own permissions are the only fence. That is the right start for one person on his own repositories, and it is the limit on how far attention can be spread: nobody leaves an agent unattended with their whole home directory. It is also why two implementers in one project collide.

## what the idea already rests on

- **Project environments.** A project already runs in an environment of its own, approved by content. A sandbox is that idea made strict.
- **A copy of the repository per piece of work.** The CodeAF review of the same day found this to be the concrete answer to parallel work, and it is the lightest sandbox there is.
- **An earlier placeholder.** A card named "container environments" stood on this board from 2026-11-06 until the trim earlier on this date. Its one hard observation stands: the real work is providing what a harness needs outside the checkout, its login and trust state, its configuration and its session store.

## what it has to settle

- **What crosses into a sandbox and what never does.** Logins are the hard case: a harness needs one, and a sandbox is where it is least safe to put one.
- **What comes back out.** A branch, a change, a transcript. An ephemeral environment that takes its conversation with it has lost the record.
- **Reaching it.** Over the overlay, with no published port, which is `central-application`'s subject and stands on `the-orchestration-framework-is-extracted`.
- **Keeping it friction-free.** The operator should not have to think about where an agent runs in order to talk to it.

## depends on

`the-orchestration-framework-is-extracted` and `central-application`. `a-workspace-that-can-be-rebuilt-elsewhere` is close kin: both need to know exactly what a conversation is made of.

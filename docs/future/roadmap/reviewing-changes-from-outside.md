---
title: reviewing changes from outside
state: researching
created: 2026-11-18
tags: [feature, spike]
---

A change that arrives from outside is reviewed against the same qualities Michael's own reviews use, by a route the contributor can see the result of, without the canon having to be handed to them.

## why

Pull request 3 arrived on 2026-11-18 with a journal entry of its own, which says: terminus came back clean "against a new `projects/mega/` rubric (24 qualities, 10 blocking) in the canon on this machine", "whose six blocking locals reconstruct the qualities Michael's reviews cited from the spec". The contributor had no access to mega's canon, so they rebuilt it from what the repository says about it, and reviewed against the rebuild.

That is a contributor doing everything `AGENTS.md` asks, with the one piece they could not get. `AGENTS.md` tells every agent to review to clean "against the canon rubric". The rubric lives in Michael's `terminus-canon` repository, with seven quality files for mega and the general ones it draws on. The mega repository names those qualities in its documents, its cards and its code comments, and holds none of them.

It also means their review and ours can disagree without anyone knowing. A reconstructed rubric will not have the amendments made in the last two weeks: the copy-once rule for a conversation's request, pi's rules, the line that lets release end a pi process, the ruling on late output.

## to settle

- **Where mega's qualities live.** In the mega repository, where a contributor finds them and a change to a rule is reviewed with the change to the code. Or in a public canon repository that mega points at. The first is simpler for one project; the second is what terminus is built around.
- **What comes with them.** Mega's seven depend on general qualities, such as the Go conventions. A contributor needs those too, or a rubric that says which apply.
- **Who may change them.** A rule that a pull request can edit is a rule a pull request can weaken. A change to a quality wants Michael's own review, whoever proposes it.
- **What the README says to a contributor.** The contributor document was left out on 2026-11-17. This is one of the things it would have to say.

## Michael's direction, 2026-11-18

"i suspect that mega might pick up a few contributors, so we'll need to keep working on figuring out how to take their changes." And on the canon: "maybe rather than publishing our canon we come up with an github action hook that allows github ci to execute terminus against our private repo?"

So the canon stays private, and the review comes to the pull request. This card was first written as "the canon for mega is not public"; it is now about that route.

## what a review in CI needs

- **The canon,** fetched from the private repository by the job, with a key that can read it and nothing else.
- **Terminus and a reviewer.** A terminus review is a coding agent reading the change against the rubric. In CI that agent needs a harness on the runner and a model to call, which means a model key held as a secret, and a cost per review.
- **Somewhere to put the verdict:** a check on the pull request, with the findings as annotations or a comment.

## what makes it dangerous, and what answers it

A pull request from a fork is untrusted text and untrusted code, and this job would hold two secrets: the key to the canon and the key to a model.

- **GitHub gives no secrets to a workflow run from a fork's pull request,** by design. The way round, running in the base repository's context, is the well-known way to leak them if that run ever executes the pull request's code. A review only needs to read the change. It must never build it, test it or run its scripts.
- **The reviewer is a model reading untrusted text.** A change can carry text written to steer it: to print the rubric, or to print what is in its environment. The job should hold no token that can write anywhere, the model key should be good for nothing else, and what the reviewer may do should be reading files and writing findings.
- **Findings quote the rule they cite.** A review posted on a public pull request shows, finding by finding, what the private qualities say. A private canon reviewed in public is private in bulk, not in substance. If that matters, the canon is not what should be kept back.
- **Who starts it.** Not every pull request automatically. A run that Michael starts, by a label or by hand with a pull request's number, keeps the secrets behind his own decision and stops a stranger spending his model budget.

## the smaller version, which works today

What was done for pull requests 2 and 3: merge to a branch or to main, run terminus here against the real canon, and fix what it finds in a follow-up. It costs Michael's attention per contribution and needs nothing built. It stops being enough at a handful of contributors.

## also to settle

- **What a contributor is told.** That the review exists, what it checks in outline, and that a reconstructed rubric is not it. The contributor document was left out on 2026-11-17; this is one of the things it would say.
- **Whether part of the canon could be public after all.** Mega's own seven qualities and the general ones they draw on are a small part of the canon repository, which holds other projects' qualities too. Publishing that part and keeping the rest is a third option between all and nothing.

---
title: remote access through zrok
state: researching
created: 2026-11-16
tags: [feature, spike]
---

Michael reaches the mega on `fortyfive` from another machine through a private zrok share, while a page stays open on `fortyfive` itself. Find out what works as built, write down the safe way to do it, and fix what gets in the way.

## why

Michael, 2026-11-16: "i am going to be going remote for most of this week, and i would like to use zrok to remotely access the mega instance on this system. will that coexist with the browser window open on this system?" The spec's answer to remote access is the central application, with a login, and that is far off. This is the interim, and he is about to use it.

## what the code says, before anyone has tried it

- **Several pages on one mega are supported by design.** The service is the only authority and every page is a client with its own stream. Stars, attention, requests and the shutdown gate are the service's and look the same everywhere. A request is answered once; a second answer is told it is no longer pending.
- **What stays in one browser:** the panel's width and folds, the reply box's height, and drafts. A draft typed on one machine is not on the other.
- **A page left open and visible reads on his behalf.** A page posts the reading position while its conversation is open and the page reports itself visible. A window left open on `fortyfive` showing a conversation will keep marking that conversation's output as read, so it never shows as unread on the remote page. Leave the local page on a project, or minimised, or closed.
- **Mega has no login.** Whoever reaches the port can launch agents, switch them to auto mode and answer approvals, as him. A share is an SSH-equivalent credential. A public share is out of the question; a private one, closed to his own account, is the least that will do.
- **The request guard may refuse a proxied page.** `internal/server/guard.go` requires the request's `Host` to be a loopback address and, when a request carries an `Origin`, that its host and port equal the `Host`. Every write and the live stream carry an `Origin`. So through a proxy the page may load and then fail on every action with "forbidden origin", depending on which `Host` the share's backend sends on.
  - **The likely way through, to be tried:** bind the remote end at the same address and port mega listens on, `127.0.0.1:8470`, so the browser's `Origin` and the `Host` mega sees agree whichever way the proxy treats the header.
  - A public share fails the guard outright, which is the right outcome.
- **The live stream is a WebSocket,** so the share has to carry one.

## to establish

- Whether the page works end to end through a private share, and what the guard needed.
- What the reconnect looks like when the share drops and returns.
- Whether anything on the page assumes the browser is on the same machine: the terminal command it offers to copy, paths in notices, the browser notification at idle.

## hazards while away

- **Shutting mega down from the remote page** leaves nothing to start it again without another way into the machine.
- **The desktop session.** The daemon runs under it. A suspend, a lock that stops the session, or a logout takes mega with it.
- **A rebuild on `fortyfive` does not restart the service.** The remote page keeps talking to the old one until someone restarts it there.

## where this leads

If the guard needs loosening to make this work, that is a decision about mega's safety posture and goes to the spec first, not into a batch. The proper answer remains a login, which `central-application` holds.

## first use, 2026-11-16

Michael tried it the same day, before leaving: "seems to work just fine. i'm using private closed sharing, bound to localhost on the accessing system." So the page works end to end through a private, closed zrok share with the remote end bound to loopback, and the request guard did not get in the way. No change to mega was needed.

Still to record from a week of real use: which port the remote end was bound to, which says whether the guard passed because the ports matched or because the proxy keeps the browser's `Host`; how a dropped share reconnects; and anything on the page that assumed the browser was on the same machine.

The hazards above are covered for this trip: Michael can also reach `fortyfive` by RDP into the desktop and by ssh, so a stopped or stale service can be restarted from away.

Later the same day Michael asked about reaching several megas at once, each share bound to its own loopback port. The page is indifferent to its port: it takes the address for its requests and its stream from wherever it was loaded. Each port is its own origin to the browser, so each keeps its own preferences. The guard still requires a request's `Origin` and `Host` to agree, which they do on any port provided the share passes the browser's `Host` through. What is missing is any sign of which machine a tab belongs to; that is the card `which-mega-is-this`.

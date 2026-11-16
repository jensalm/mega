---
title: which mega is this
state: researching
created: 2026-11-16
tags: [enhancement]
---

A page says which machine's mega it is showing, in its tab title and in the side panel, so two megas open side by side cannot be mistaken for each other.

## why

Michael, 2026-11-16, working away from his desk: he reaches the megas on his hosts through private zrok shares, each bound to its own loopback port on the machine he is using. The page takes its address from wherever it was loaded, so the port does not matter to it, and each port keeps its own browser preferences. But nothing on the page names the machine. Two tabs are both titled "mega" with the same mark and the same layout, and a prompt or an approval meant for one host can go to the other.

## shape

- The service says what it is called: the machine's hostname by default, or a name set in mega's config file.
- The tab title carries it, as in "mega · fortyfive", and keeps it through the titles the page already sets, such as the one for a shutdown that is ready.
- The side panel's footer shows it beside the connection status, and the rail carries it as a title.
- The version panel and `mega version` say it too.
- The hostname is the operator's own machine's name and is shown only to someone who can already reach the page. It goes in no log line that leaves the machine and no committed fixture.

## checks

In the side suite: the name in the tab title, the footer and the rail, with a name set in the config and with the default.

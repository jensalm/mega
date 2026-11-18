---
title: improve the new project dialog
state: evaluating
created: 2026-11-18
tags: [enhancement]
---

could we have a file picker?

## background

Noted by the planning agent on 2026-11-18. The dialog today is two text fields: an absolute path typed by hand, and a name.

**The browser's own file picker cannot do this,** for two reasons:

- A browser never tells a page the real path of what was picked. It hands over the contents, or a handle, and a folder's name at most. Registering a project needs the path.
- The picker shows the machine the browser is on. Michael reaches megas on other machines, and a project's checkout is on the service's machine.

So the picker has to be mega's own, drawn in the page and answered by the service, as the reply box's path search is.

**A shape to react to:**

- The checkout field completes as it is typed, from the service: `~/Rep` offers the folders that match, and `~` is accepted for the home directory.
- Below it, a list of the folder being looked at, to click through, with a way up. Folders only, names only.
- A folder that is a git checkout is marked, and one already registered says so and cannot be picked twice.
- The name fills in from the folder and can be changed.

**To settle.** How much of the machine the list may show. The reply box's search lists nothing outside directories the operator named, because what it finds is one keystroke from an agent. This list only ever leads to registering a project, but it does show folder names to whoever can reach the page. Starting at the home directory, with roots that can be named in the config, is the cautious version.

## settled with Michael, 2026-11-18

The picker starts at the home directory, or at a projects root named in the config, `~/Repos` in his case, and shows nothing above it. A path typed in full is still accepted wherever it is, as now; only what the picker lists and completes is confined to the root.

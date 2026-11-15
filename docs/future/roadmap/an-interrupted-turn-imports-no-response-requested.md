---
title: an interrupted turn imports "no response requested"
state: researching
created: 2026-11-14
tags: [defect]
---

After a turn is interrupted, the conversation released and then resumed, the history gains one stray line from the agent: "No response requested." Claude wrote it to its own transcript for the interrupted turn. It is Claude's bookkeeping, not speech, and should not appear as a reply.

## why

Found by the version spike on 2026-11-14, on every Claude Code build from 2.1.318 to 2.1.325, so it is not a version difference. The line never arrives on the live stream. It appears only when the resume imports native history, which reads the transcript and finds a message mega has not seen. The same shape as the "outside mega" block of 2026-11-10, where the history reader delivered messages the live stream never did, and it was fixed in the adapter's mapping.

The round trips do not catch it, because none interrupts before its import step.

## shape

- The Claude adapter's mapping treats this synthetic message as a native event, as it does the messages marked `is_meta`. Establish how the transcript marks it, so the rule is not a match on the sentence.
- Whether an import of it counts for attention. Under `bookkeeping-never-creates-attention` it must not.
- An e2e round trip that interrupts, releases and resumes, and asserts the import brings in no reply. A Go test over a recorded transcript for the mapping.

import type { Conversation } from './types'

// lastInstructed is, for each project, the conversation the operator most
// recently instructed: the one whose newest prompt sent through mega is the
// latest (instructed_at, kept by the service). a project where nothing was
// instructed has none. it is a fact about the operator, never attention: an
// agent's own work never moves it, and it orders and counts nothing. when
// the marked conversation leaves the workspace, the next most recent in its
// project takes the mark, since it is reckoned from what remains.
export function lastInstructed(conversations: Conversation[]): Set<string> {
  const newest = new Map<string, { id: string; at: number }>()
  for (const conversation of conversations) {
    const at = Date.parse(conversation.instructed_at ?? '')
    if (Number.isNaN(at)) {
      continue
    }
    const held = newest.get(conversation.project_id)
    if (!held || at > held.at) {
      newest.set(conversation.project_id, { id: conversation.id, at })
    }
  }
  return new Set([...newest.values()].map((held) => held.id))
}

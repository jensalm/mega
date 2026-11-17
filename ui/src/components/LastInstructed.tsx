import { shortTime } from '../zone'
import type { Conversation } from '../types'
import { Icon } from './Icon'

// LastInstructed is the small muted arrow after the name of the conversation
// the operator last instructed in its project (lastInstructed), pointing back
// at it: where he left off. its title says when, in the service's zone, with the date when
// it was another day. it is not attention, and takes no mark from that set.
export function LastInstructed({ conversation }: { conversation: Conversation }) {
  if (!conversation.instructed_at) {
    return null
  }
  const label = `You last instructed this conversation in its project, at ${shortTime(new Date(conversation.instructed_at))}`
  return (
    <span className="last-instructed" role="img" title={label} aria-label={label}>
      <Icon name="west" />
    </span>
  )
}

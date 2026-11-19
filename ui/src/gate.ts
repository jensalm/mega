import type { Blocker, ShutdownGate } from './types'

// pendingReason is why new work waits while the gate is armed: the title of
// every control it holds back.
export const pendingReason = 'Waiting to shut down once every agent is idle; cancel the shutdown to start new work'

// shutDownReason is why new work waits on the page that shut the service down,
// until the service is back.
export const shutDownReason = 'mega was shut down from this page; new work waits for the service to start again'

// shutDownLostReason is why new work waits on a page that asked for a
// shutdown and then lost contact before mega said it had finished.
export const shutDownLostReason = 'A shutdown was asked for from this page and contact was lost; new work waits until mega answers again'

// shutDownFooter is what the footer and the rail say while a page that shut
// mega down is out of touch: shut down, when it heard mega finish, or asked
// for, when it lost contact first.
export const shutDownFooter = (heard: boolean) => (heard ? 'Shut down from here, retrying' : 'Shutdown asked, contact lost, retrying')

const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`

// blockerKinds names what one conversation holds the gate with. a request is
// the operator's to answer, since nothing else will clear it; the turn it
// holds open is not named again beside it.
export function blockerKinds(b: Blocker): string[] {
  const kinds: string[] = []
  if (b.requests > 0) {
    kinds.push(b.requests === 1 ? 'a request for you to answer' : `${b.requests} requests for you to answer`)
  }
  if (b.turn && b.requests === 0) {
    kinds.push('a turn in flight')
  }
  if (b.background > 0) {
    kinds.push(plural(b.background, 'background job', 'background jobs'))
  }
  if (b.settling) {
    kinds.push(b.settling)
  }
  return kinds
}

// gateCounts is what the gate waits for, by kind, across the workspace.
export function gateCounts(gate: ShutdownGate): string {
  let turns = 0
  let requests = 0
  let background = 0
  let settling = 0
  for (const b of gate.blockers) {
    turns += b.turn && b.requests === 0 ? 1 : 0
    requests += b.requests
    background += b.background
    settling += b.settling ? 1 : 0
  }
  const parts = [
    requests > 0 && plural(requests, 'request for you to answer', 'requests for you to answer'),
    turns > 0 && plural(turns, 'turn in flight', 'turns in flight'),
    background > 0 && plural(background, 'background job', 'background jobs'),
    settling > 0 && plural(settling, 'conversation starting', 'conversations starting'),
  ].filter(Boolean)
  return parts.join(' · ')
}

// gateSummary is the gate in one line, for a title or a label.
export function gateSummary(gate: ShutdownGate): string {
  if (gate.idle) {
    return 'Every agent is idle: ready to shut down'
  }
  return `Waiting to shut down: ${gate.blockers.map((b) => `${b.name} (${blockerKinds(b).join(', ')})`).join('; ')}`
}

// clockTime is a time of day as the page shows it, in the service's zone.
export { clockTime } from './zone'

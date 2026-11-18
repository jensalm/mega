// writes a long synthetic pi session file, for timing a read of one: where
// pi keeps it under an agent directory (PI_CODING_AGENT_DIR), in the shape
// pi 1.5.2 writes, each turn a prompt, a tool call, its result and a reply,
// of filler text. kept here, beside the fake pi, so pi's protocol stays with
// its adapter; probes/look/gen.mjs calls it. it returns the session's id.
import { mkdirSync, writeFileSync } from 'node:fs'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

const filler = (n) => 'filler text standing in for a long reply. '.repeat(n)

export function writeLongSession(agentDir, cwd, turns) {
  const piID = randomUUID()
  const sessions = join(agentDir, 'sessions', `--${cwd.replace(/^\//, '').replace(/[/\\:]/g, '-')}--`)
  mkdirSync(sessions, { recursive: true })
  let at = Date.parse('2026-11-18T09:00:00Z')
  const pi = [JSON.stringify({ type: 'session', version: 3, id: piID, timestamp: new Date(at).toISOString(), cwd })]
  let piParent = null
  const piEntry = (message) => {
    const id = randomUUID().slice(0, 8)
    pi.push(JSON.stringify({ type: 'message', id, parentId: piParent, timestamp: new Date((at += 1000)).toISOString(), message: { ...message, timestamp: at } }))
    piParent = id
  }
  for (let i = 1; i <= turns; i++) {
    piEntry({ role: 'user', content: `turn ${i}: ${filler(4)}` })
    const call = `call_${i}`
    piEntry({ role: 'assistant', content: [{ type: 'toolCall', id: call, name: 'bash', arguments: { command: `echo ${i}` } }], stopReason: 'toolUse' })
    piEntry({ role: 'toolResult', toolCallId: call, toolName: 'bash', content: [{ type: 'text', text: filler(10) }], isError: false })
    piEntry({ role: 'assistant', content: [{ type: 'text', text: `reply ${i}: ${filler(20)}` }], stopReason: 'stop' })
  }
  writeFileSync(join(sessions, `2026-11-18T09-00-00-000Z_${piID}.jsonl`), pi.join('\n') + '\n')
  return piID
}

// writes a long synthetic Codex rollout, for timing a read of one: where
// codex keeps it under CODEX_HOME, in the shape codex-cli 0.166.0 writes,
// the session's meta, then for each turn its start, the user's message, the
// agent's and its completion, of filler text. kept here, beside the fake
// codex, so Codex's protocol stays with its adapter; probes/look/gen.mjs
// calls it. it returns the thread's id.
import { mkdirSync, writeFileSync } from 'node:fs'
import { randomUUID } from 'node:crypto'
import { join } from 'node:path'

const filler = (n) => 'filler text standing in for a long reply. '.repeat(n)

export function writeLongRollout(codexHome, cwd, turns) {
  const codexID = randomUUID()
  const rollouts = join(codexHome, 'sessions', '2026', '11', '18')
  mkdirSync(rollouts, { recursive: true })
  let t = Date.parse('2026-11-18T09:00:00Z')
  const stamp = () => new Date((t += 1000)).toISOString()
  const rollout = [JSON.stringify({ timestamp: stamp(), type: 'session_meta', payload: { id: codexID, timestamp: stamp(), cwd, originator: 'codex_cli_rs', cli_version: '0.166.0', source: 'cli', model_provider: 'openai' } })]
  for (let i = 1; i <= turns; i++) {
    const turn = `turn-${i}`
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'event_msg', payload: { type: 'task_started', turn_id: turn } }))
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'response_item', payload: { type: 'message', role: 'user', content: [{ type: 'input_text', text: `turn ${i}: ${filler(4)}` }] } }))
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'event_msg', payload: { type: 'user_message', message: `turn ${i}: ${filler(4)}`, images: [] } }))
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'response_item', payload: { type: 'message', role: 'assistant', content: [{ type: 'output_text', text: `reply ${i}: ${filler(20)}` }] } }))
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'event_msg', payload: { type: 'agent_message', message: `reply ${i}: ${filler(20)}` } }))
    rollout.push(JSON.stringify({ timestamp: stamp(), type: 'event_msg', payload: { type: 'task_complete', turn_id: turn, last_agent_message: `reply ${i}` } }))
  }
  writeFileSync(join(rollouts, `rollout-2026-11-18T09-00-00-${codexID}.jsonl`), rollout.join('\n') + '\n')
  return codexID
}

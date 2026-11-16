#!/usr/bin/env node
// outside-the-checkout probe (2026-11-16): what a codex agent does with a
// file named outside its checkout, as the reply box's path search would
// insert it: a bare ~/dir/file against a bare absolute path, and a name with
// spaces. and what codex asks before reading there.
//
//   node outside.mjs --cwd DIR --file /abs/dir/file [--spaced "/abs/dir/a b"] --log logs/x.jsonl
//
// one thread, one turn per form, each asking for the file's exact contents.
// every approval is declined and recorded, so a read that needs one fails
// and says so. records the commands run, the approvals asked and the reply.
import { spawn } from 'node:child_process'
import { createWriteStream, readFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { resolve } from 'node:path'
import { createInterface } from 'node:readline'

const argv = process.argv.slice(2)
const opt = {}
for (let i = 0; i < argv.length; i++) {
  if (!argv[i].startsWith('--')) continue
  const next = argv[i + 1]
  if (next === undefined || next.startsWith('--')) opt[argv[i].slice(2)] = true
  else { opt[argv[i].slice(2)] = next; i++ }
}
const cwd = resolve(opt.cwd ?? '.')
const tilde = (p) => (p.startsWith(homedir() + '/') ? '~' + p.slice(homedir().length) : p)
const forms = [
  ['tilde', opt.file, tilde(opt.file)],
  ['absolute', opt.file, opt.file],
]
if (typeof opt.spaced === 'string') {
  forms.push(['spaced tilde', opt.spaced, tilde(opt.spaced)], ['spaced absolute', opt.spaced, opt.spaced])
}
const log = createWriteStream(opt.log ?? `logs/outside-${Date.now()}.jsonl`)
const t0 = Date.now()
const rec = (kind, data) => log.write(JSON.stringify({ ms: Date.now() - t0, kind, ...data }) + '\n')
const child = spawn('codex', ['app-server'], { cwd, stdio: ['pipe', 'pipe', 'pipe'] })
let nextId = 1
const waiting = new Map()
const send = (frame) => { rec('send', { frame }); child.stdin.write(JSON.stringify(frame) + '\n') }
const request = (method, params) => new Promise((res, rej) => { const id = nextId++; waiting.set(id, { res, rej }); send({ id, method, params }) })
let turn = null
const turns = []
createInterface({ input: child.stdout }).on('line', (line) => {
  if (!line.trim()) return
  const msg = JSON.parse(line)
  rec('recv', { frame: msg })
  if (msg.id !== undefined && msg.method === undefined) {
    const w = waiting.get(msg.id)
    if (w) { waiting.delete(msg.id); msg.error ? w.rej(new Error(JSON.stringify(msg.error))) : w.res(msg.result) }
    return
  }
  if (msg.id !== undefined && msg.method) {
    turn?.asked.push({ method: msg.method, command: msg.params?.command ?? null, reason: msg.params?.reason ?? null })
    send({ id: msg.id, result: { decision: 'decline' } })
    return
  }
  const it = msg.params?.item
  if (turn && msg.method === 'item/completed' && it) {
    if (it.type === 'commandExecution') turn.commands.push({ command: it.command, status: it.status, exitCode: it.exitCode ?? null })
    if (it.type === 'agentMessage') turn.reply = it.text
  }
  if (turn && msg.method === 'turn/completed') turn.done = true
})
await request('initialize', { clientInfo: { name: 'mega-probe', version: '0' }, capabilities: { experimentalApi: true } })
send({ method: 'initialized', params: {} })
const thread = await request('thread/start', { cwd })
rec('thread', { approvalPolicy: thread.approvalPolicy ?? null, sandbox: thread.sandbox ?? null })
for (const [name, path, named] of forms) {
  const content = readFileSync(path, 'utf8').trim()
  turn = { name, named, commands: [], asked: [], reply: null, done: false }
  turns.push(turn)
  await request('turn/start', { threadId: thread.thread.id, input: [{ type: 'text', text: `Reply with the exact contents of the file ${named} and nothing else. If you cannot read it, reply NOT-READ.` }] })
  while (!turn.done) await new Promise((r) => setTimeout(r, 200))
  turn.read = turn.reply?.includes(content) ?? false
}
child.stdin.end()
await new Promise((r) => child.on('exit', r))
const summary = { thread: thread.thread.id, approvalPolicy: thread.approvalPolicy ?? null, sandbox: thread.sandbox ?? null, turns }
rec('summary', summary)
log.end()
console.log(JSON.stringify(summary, null, 2))

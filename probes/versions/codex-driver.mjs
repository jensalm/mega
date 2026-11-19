// the Codex version driver: every call mega makes of Codex, made over a
// running scratch mega's HTTP API, against the build its config names in
// codex.binary, in a scratch CODEX_HOME. it mirrors the e2e round trips
// (internal/e2e/codex_test.go).
//
//   MEGA_PROBE_URL=http://127.0.0.1:PORT MEGA_E2E_CODEX_FIXTURE=... \
//     MEGA_PROBE_CODEX_HOME=/abs/scratch/codex-home node codex-driver.mjs [--setup-only]
//   ... node codex-driver.mjs --adopt THREAD [--setup-only]
//
// run it through codex.sh, which builds the scratch CODEX_HOME and starts and
// stops the scratch mega. nothing runs Codex until the project's own
// resolution from the override is seen to set CODEX_HOME to the scratch home.
// it answers exactly two approvals, each asserted before Allow: the shell
// write outside the checkout, and the fixture's MCP echo. the interrupted
// approval is never answered. --adopt is the second mega's run, on fresh
// data and the same scratch home: discovery lists THREAD and adopting it
// imports its history. --setup-only stops after the guard.
import { existsSync, readFileSync, rmSync } from 'node:fs'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import { cannotRun, flag, option, probe, sleep } from './lib.mjs'

const markerPrompt = 'Reply with the project instruction marker token from your project instructions (AGENTS.md), followed by the word ALPHA-1a2b. Nothing else.'
// the fixture is trusted in the operator's codex config, which gives it a
// workspace-write sandbox, so only a write outside the checkout needs approval.
const approvalCommand = 'echo GAMMA-5e6f > ../codex-rt1-approval.txt'
const approvalPrompt = `Run exactly this shell command, and no other: ${approvalCommand} (it writes in the parent of the working directory). Then reply with GAMMA-5e6f. Nothing else.`
const interruptCommand = 'echo EPSILON-2b9c > ../codex-rt1-interrupt.txt'
const interruptPrompt = `Run exactly this shell command, and no other: ${interruptCommand} (it writes in the parent of the working directory). Then reply with EPSILON-2b9c. Nothing else.`
const mcpPrompt = 'Call the echo tool from the fixture MCP server with the text MCP-E2E-7c1d, then reply with exactly what the tool returned. Nothing else.'
const startPrompt = 'Reply with the word BETA-START-9e1f. Do not run any commands. Nothing else.'
const switchPrompt = 'Reply with the word BETA-SWITCH-4c2d. Do not run any commands. Nothing else.'
const returnPrompt = 'List every marker token you have seen in this conversation so far, in order, one per line, then append RESUME-SDK-c3d4 on its own line. Nothing else.'

const fixture = process.env.MEGA_E2E_CODEX_FIXTURE
if (!fixture) cannotRun('MEGA_E2E_CODEX_FIXTURE is not set: the probe needs the disposable codex-rt1 fixture (see internal/e2e/doc.go)')
const codexHome = process.env.MEGA_PROBE_CODEX_HOME
if (!codexHome || !isAbsolute(codexHome)) cannotRun('MEGA_PROBE_CODEX_HOME must name the scratch CODEX_HOME, as an absolute path')
if (resolve(codexHome) === resolve(process.env.HOME ?? '/nonexistent', '.codex')) cannotRun('MEGA_PROBE_CODEX_HOME is the real ~/.codex; the probe runs Codex only in a scratch home')
const adopt = option('--adopt')
if (flag('--adopt') && !adopt) cannotRun('--adopt needs the thread to adopt')

// what the probe writes and reads in the disposable fixture: the approved
// command's file and the interrupted one's, beside the checkout, and the
// fixture MCP server's call log.
const approvalFile = join(dirname(resolve(fixture)), 'codex-rt1-approval.txt')
const interruptFile = join(dirname(resolve(fixture)), 'codex-rt1-interrupt.txt')
const mcpCalls = join(fixture, 'mcp', 'mcp-calls.log')

// single quotes, so the path reaches the shell as it is.
const quoted = `'${codexHome.replaceAll("'", `'\\''`)}'`
const activation = `export CODEX_HOME=${quoted}`

const p = probe('codex')
const { summary, check, must } = p

// codex reports a command inside the login shell it runs it in
// ("/bin/bash -lc '...'"); the command itself is what is compared.
const unwrapShell = (command) => {
  const m = /^\S+ -lc '(.*)'$/s.exec(command ?? '')
  return m ? m[1] : command
}
const shell = (want) => (payload) => {
  if (payload.tool !== 'shell') return `${payload.tool}`
  const got = unwrapShell(payload.input?.command)
  return got === want ? '' : `shell \`${payload.input?.command}\``
}
const echo = (payload) => {
  if (payload.tool !== 'fixture/echo' || payload.mcp_server !== 'fixture' || payload.mcp_tool !== 'echo') return `${payload.tool} (${payload.mcp_server}/${payload.mcp_tool})`
  if (payload.input?.text !== 'MCP-E2E-7c1d') return `fixture/echo of ${JSON.stringify(payload.input)}`
  if (payload.allow_session !== true) return 'fixture/echo without the offer for the session'
  return ''
}

// guard is the gate before anything runs Codex: the project's own resolution
// from the override, after the override was set, changes CODEX_HOME, and a
// live resolution sets it to the scratch home. it reads that one value and no
// other.
const guard = async (projectId, overrideAt) => {
  const resolved = await p.fromOverride(projectId, overrideAt)
  must(resolved.changed.includes('CODEX_HOME'), `the override's resolution changes CODEX_HOME (changed: ${resolved.changed.join(', ') || 'nothing'})`)
  const live = await p.resolvedValues(projectId, ['CODEX_HOME'])
  must(live.source === 'override' && live.values.CODEX_HOME === codexHome, `a live resolution, from the ${live.source ?? 'no'} source, sets CODEX_HOME to the scratch home ${codexHome}${live.error ? `: ${live.error}` : ''}`)
  summary.codex_home = live.values.CODEX_HOME
  summary.environment = { source: resolved.source, changed: resolved.changed }
}

// the run in the second mega: discovery and adoption.
const adoption = async () => {
  const { project, overrideAt } = await p.register('codex-rt1', fixture, activation)
  await guard(project.id, overrideAt)
  if (flag('--setup-only')) {
    check(true, 'setup only: stopping before anything runs Codex')
    return
  }
  const found = await p.api('GET', `/api/projects/${project.id}/discover?harness=codex`)
  must(found.status === 200, `discovery answered${found.status === 200 ? '' : `: ${found.status} ${JSON.stringify(found.body)}`}`)
  const sessions = found.body.sessions ?? []
  summary.discovered = sessions.length
  must(sessions.some((s) => s.native_id === adopt), `discovery lists thread ${adopt}, among ${sessions.length}`)
  const r = await p.api('POST', `/api/projects/${project.id}/adopt`, { harness: 'codex', native_id: adopt, name: 'versions probe, adopted' })
  must(r.status === 201, `adopted thread ${adopt}${r.status === 201 ? '' : `: ${r.status} ${JSON.stringify(r.body)}`}`)
  const evs = await p.events(r.body.id)
  const counts = {}
  for (const e of evs) if (e.origin === 'native') counts[e.kind] = (counts[e.kind] ?? 0) + 1
  const resumed = evs.filter((e) => e.kind === 'runtime' && e.status === 'resumed').pop()
  check((counts.user ?? 0) > 0 && (counts.assistant ?? 0) > 0, `adopting imported its history: ${resumed?.text ?? 'no resumed line'} (${Object.entries(counts).map(([k, v]) => `${k} ${v}`).join(', ')})`)
  summary.adopted = { conversation: r.body.id, native_id: adopt, imported: counts }
}

// the main run.
const roundTrip = async () => {
  await p.startup('codex')
  const { project, overrideAt } = await p.register('codex-rt1', fixture, activation)
  await guard(project.id, overrideAt)
  if (flag('--setup-only')) {
    check(true, 'setup only: stopping before anything runs Codex')
    return
  }

  // the model list, and what a new conversation copies.
  const t0 = Date.now()
  const models = await p.api('GET', `/api/projects/${project.id}/models?harness=codex`)
  const ms = Date.now() - t0
  must(models.status === 200, `the model list answered${models.status === 200 ? '' : `: ${models.status} ${JSON.stringify(models.body)}`}`)
  const copy = models.body.copy ?? {}
  summary.models = { count: models.body.models.length, version: models.body.version, ms, copy }
  check(models.body.models.length > 0 && !!models.body.version, `the model list: ${models.body.models.length} models, at version ${models.body.version}, in ${(ms / 1000).toFixed(1)} s; a new conversation copies ${copy.model} at ${copy.effort}`)

  // a launch and the marker turn, in the scratch home.
  const conv = await p.launch(project.id, { harness: 'codex', name: 'versions probe', prompt: markerPrompt })
  const id = conv.id
  await p.turnEnd(id, 0, 'the marker turn')
  await p.expectReply(id, 'ROUNDHOUSE-MARKER-7f3a', 'ALPHA-1a2b')
  const c = await p.convOf(id)
  const thread = c.native_id
  summary.conversation = id
  summary.native_id = thread
  check(!!thread && (c.permission_mode ?? '').includes('on-request') && !!c.version && !!c.model,
    `the started facts: ${c.version} running ${c.model} at ${c.effort}, policy ${c.permission_mode}, thread ${thread}`)
  check(c.environment?.source === 'override' && (c.environment?.changed ?? []).includes('CODEX_HOME'),
    `the runtime started in the override's environment (${c.environment?.source}; changed ${(c.environment?.changed ?? []).join(', ')})`)

  // an approval for a write outside the checkout, asserted before Allow.
  rmSync(approvalFile, { force: true })
  let seq = await p.lastSeq(id)
  await p.prompt(id, approvalPrompt)
  const first = await p.approval(id, seq, `shell \`${approvalCommand}\``, shell(approvalCommand))
  await sleep(4000)
  check((await p.convOf(id)).turn_active, 'the turn waits while the approval is pending')
  must(!existsSync(approvalFile), 'nothing was written before Allow')
  must((await p.answer(id, first.id, 'allow')) === '', 'Allow was accepted')
  await p.turnEnd(id, seq, 'the approved turn')
  await p.expectReply(id, 'GAMMA-5e6f')
  check(existsSync(approvalFile) && readFileSync(approvalFile, 'utf8').includes('GAMMA-5e6f'), 'the approved command wrote its file')

  // an interrupt with a second approval pending: never answered.
  rmSync(interruptFile, { force: true })
  seq = await p.lastSeq(id)
  await p.prompt(id, interruptPrompt)
  const second = await p.approval(id, seq, `shell \`${interruptCommand}\``, shell(interruptCommand))
  await p.interrupt(id)
  await p.interrupted(id, seq, second.id)
  await sleep(2000)
  check(!existsSync(interruptFile), 'the interrupted command wrote nothing')

  // the MCP approval path: not run before Allow, run after.
  rmSync(mcpCalls, { force: true })
  seq = await p.lastSeq(id)
  await p.prompt(id, mcpPrompt)
  const tool = await p.approval(id, seq, 'fixture/echo of MCP-E2E-7c1d, offered for the session', echo)
  await sleep(3000)
  must(!existsSync(mcpCalls), 'the MCP tool did not run before Allow')
  must((await p.answer(id, tool.id, 'allow')) === '', 'Allow was accepted')
  await p.turnEnd(id, seq, 'the MCP turn')
  await p.expectReply(id, 'echo: MCP-E2E-7c1d')
  check(existsSync(mcpCalls) && readFileSync(mcpCalls, 'utf8').includes('MCP-E2E-7c1d'), 'the fixture server logged the call after Allow')
  rmSync(mcpCalls, { force: true })

  // MCP status, by counts.
  summary.mcp = await p.mcpCounts(id)

  // a release, then a start without a prompt on the same thread.
  await p.release(id)
  const idle = await p.startIdle(id)
  check(idle.view.model === idle.view.model_choice, `the idle start runs the conversation's model (${idle.view.model}, chosen ${idle.view.model_choice})`)
  summary.idle_start = { version: idle.view.version, model: idle.view.model, effort: idle.view.effort }
  summary.imports = { start: idle.imports }

  // a prompt answered in the resumed thread.
  seq = await p.lastSeq(id)
  await p.prompt(id, startPrompt)
  await p.turnEnd(id, seq, 'the prompt in the resumed thread')
  await p.expectReply(id, 'BETA-START-9e1f')
  await p.sameSession(id, thread, 'the prompt after the start')

  // a model switch while idle: codex takes a model choice with the next
  // prompt, so it is stored for it (next_prompt), and that turn must run on it.
  summary.idle_switch = await p.idleSwitch(id, models.body.models, { prefer: 'gpt-6-luna', outcome: 'next_prompt', ask: switchPrompt, reply: 'BETA-SWITCH-4c2d' })

  // a release, and an exact-id resume by a prompt that must remember.
  await p.release(id)
  seq = await p.lastSeq(id)
  await p.prompt(id, returnPrompt)
  await p.turnEnd(id, seq, 'the resumed turn')
  await p.expectReply(id, 'ROUNDHOUSE-MARKER-7f3a', 'ALPHA-1a2b', 'GAMMA-5e6f', 'RESUME-SDK-c3d4')
  await p.sameSession(id, thread, 'the exact-id resume')
  const resumeImports = await p.imports(id, seq)
  check(resumeImports.length === 0, `the exact-id resume imported nothing${resumeImports.length ? `: ${resumeImports.map((e) => `${e.kind} "${e.text}"`).join('; ')}` : ''}`)
  summary.imports.resume = resumeImports

  // every turn's line.
  summary.turns = await p.turns(id)
  summary.versions = await p.started(id)
  summary.version = summary.versions.length === 1 ? summary.versions[0] : summary.versions
  check(summary.versions.length === 1, `every start reported one version: ${summary.versions.join(', ')}`)

  await p.release(id)
  // the approved command's file is the probe's own.
  rmSync(approvalFile, { force: true })
}

await p.run(adopt ? adoption : roundTrip)

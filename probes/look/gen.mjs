#!/usr/bin/env node
// writes a long synthetic session for each harness, for the adopt dialog's
// look to read, in a scratch directory, never a real store: a Claude Code
// transcript under SCRATCH/claude (CLAUDE_CONFIG_DIR), a Codex rollout under
// SCRATCH/codex (CODEX_HOME) and a pi session file under SCRATCH/pi
// (PI_CODING_AGENT_DIR), each of TURNS turns of filler. each harness's
// format is written by a writer kept beside its fake, so the protocols stay
// with the adapters and the shim. it prints the three ids.
//
//   node gen.mjs SCRATCH CWD TURNS
import { join } from 'node:path'
import { writeLongTranscript } from '../../shim/claude/fake/transcript.mjs'
import { writeLongRollout } from '../../internal/harness/codex/fake/rollout.mjs'
import { writeLongSession } from '../../internal/harness/pi/fake/session.mjs'

const [scratch, cwd, turnsArg] = process.argv.slice(2)
const turns = Number(turnsArg) || 2000
console.log(
  JSON.stringify({
    claude: writeLongTranscript(join(scratch, 'claude'), cwd, turns),
    codex: writeLongRollout(join(scratch, 'codex'), cwd, turns),
    pi: writeLongSession(join(scratch, 'pi'), cwd, turns),
  }),
)

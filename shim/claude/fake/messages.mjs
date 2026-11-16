// the SDK messages the fake shim sends for a tool call, shaped as claude
// code's: an assistant message holding the tool_use block, and a user
// message holding its tool_result. kept here, beside the fake shim, so the
// Claude protocol stays with the shim; the fake shim sends them, and a check
// that writes a scratch history (checks/lib/scene.mjs) records them as the
// adapter would.

// toolUseMessage is the assistant message that calls tool name, with input,
// under the block id.
export const toolUseMessage = ({ id, name, input, uuid, session }) => ({
  type: 'assistant',
  uuid,
  session_id: session,
  parent_tool_use_id: null,
  message: { role: 'assistant', content: [{ type: 'tool_use', id, name, input }] },
})

// toolResultMessage is the user message carrying the result of the call id,
// marked is_error when it failed.
export const toolResultMessage = ({ id, result, isError = false, uuid, session }) => ({
  type: 'user',
  uuid,
  session_id: session,
  parent_tool_use_id: null,
  message: { role: 'user', content: [{ type: 'tool_result', tool_use_id: id, content: result, ...(isError ? { is_error: true } : {}) }] },
})

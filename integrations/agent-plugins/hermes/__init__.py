"""koment plugin for Hermes Agent.

Wires two hooks onto the koment CLI:

* ``pre_tool_call`` — before Hermes writes or patches a file, the content is
  handed to ``koment agents hook pre-tool``. When the edit adds ordinary
  explanatory comment intent the write is denied and the reason tells the agent
  to record the rationale as an annotation instead.

* ``pre_verify`` — before a turn that edited code is allowed to finish,
  ``koment agents hook stop`` runs the three repository gates. A failure keeps
  the agent working rather than stopping with the repository in a state its own
  policy rejects.

Both hooks shell out to the ``koment`` binary, which must be on ``PATH``. The
decision logic lives in Go so that this plugin, the Claude Code hooks, the
OpenCode plugin and CI cannot disagree about what the policy is.

Set ``KOMENT_PLUGIN_DISABLED=1`` to make both hooks no-ops without uninstalling.
"""

from __future__ import annotations

import json
import logging
import os
import shutil
import subprocess
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)

KOMENT_BINARY = "koment"
PRE_TOOL_TIMEOUT_SECONDS = 10
VERIFY_TIMEOUT_SECONDS = 120

_WRITE_TOOLS: Dict[str, tuple] = {
    "write_file": ("path", ("content",)),
    "patch": ("path", ("new_string", "patch")),
}


def _disabled() -> bool:
    return os.environ.get("KOMENT_PLUGIN_DISABLED", "").strip() not in ("", "0", "false")


def _koment_available() -> bool:
    return shutil.which(KOMENT_BINARY) is not None


def _run(arguments, payload: str, timeout: int) -> Optional[subprocess.CompletedProcess]:
    try:
        return subprocess.run(
            [KOMENT_BINARY, *arguments],
            input=payload,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )
    except (OSError, subprocess.SubprocessError) as error:
        logger.warning("koment %s did not run: %s", " ".join(arguments), error)
        return None


def _edit_payload(tool_name: str, arguments: Any) -> Optional[str]:
    mapping = _WRITE_TOOLS.get(tool_name)
    if mapping is None or not isinstance(arguments, dict):
        return None
    path_key, content_keys = mapping
    path = arguments.get(path_key)
    if not isinstance(path, str) or not path:
        return None
    for key in content_keys:
        content = arguments.get(key)
        if isinstance(content, str) and content:
            return json.dumps(
                {
                    "tool_name": "opencode_edit",
                    "tool_input": {"filePath": path, "content": content},
                }
            )
    return None


def _denial_reason(stdout: str) -> Optional[str]:
    try:
        decoded = json.loads(stdout or "{}")
    except json.JSONDecodeError:
        logger.warning("koment pre-tool returned output that is not JSON")
        return None
    specific = decoded.get("hookSpecificOutput")
    if not isinstance(specific, dict):
        return None
    if specific.get("permissionDecision") != "deny":
        return None
    reason = specific.get("permissionDecisionReason")
    return reason if isinstance(reason, str) and reason else "koment policy denied this edit"


def _on_pre_tool_call(tool_name: str = "", args: Any = None, **_: Any) -> Optional[Dict[str, Any]]:
    if _disabled() or not _koment_available():
        return None
    payload = _edit_payload(tool_name, args)
    if payload is None:
        return None
    finished = _run(["agents", "hook", "pre-tool"], payload, PRE_TOOL_TIMEOUT_SECONDS)
    if finished is None:
        return None
    reason = _denial_reason(finished.stdout)
    if reason is None:
        return None
    return {"action": "block", "message": reason}


def _verification_failure() -> Optional[str]:
    finished = _run(["agents", "hook", "stop"], "{}", VERIFY_TIMEOUT_SECONDS)
    if finished is None:
        return None
    try:
        decoded = json.loads(finished.stdout or "{}")
    except json.JSONDecodeError:
        logger.warning("koment stop hook returned output that is not JSON")
        return None
    for key in ("reason", "stopReason", "systemMessage"):
        value = decoded.get(key)
        if isinstance(value, str) and value:
            return value
    return None


def _on_pre_verify(**_: Any) -> Optional[Dict[str, Any]]:
    if _disabled() or not _koment_available():
        return None
    reason = _verification_failure()
    if reason is None:
        return None
    return {"decision": "block", "reason": reason}


def register(ctx) -> None:
    ctx.register_hook("pre_tool_call", _on_pre_tool_call)
    ctx.register_hook("pre_verify", _on_pre_verify)

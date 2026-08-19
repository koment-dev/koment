use zed_extension_api::{
    self as zed,
    settings::{ContextServerSettings, LspSettings},
    ContextServerId, LanguageServerId, Project, Result, Worktree,
};

const BINARY: &str = "koment";

struct KomentExtension {
    found: Option<String>,
}

impl KomentExtension {
    fn configured_for_language_server(
        id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Option<String> {
        LspSettings::for_worktree(id.as_ref(), worktree)
            .ok()?
            .binary?
            .path
    }

    fn configured_for_context_server(id: &ContextServerId, project: &Project) -> Option<String> {
        ContextServerSettings::for_project(id.as_ref(), project)
            .ok()?
            .command?
            .path
    }
}

fn nowhere_to_be_found() -> String {
    format!(
        "koment is not on $PATH. Install it, or set the {BINARY} binary path in your Zed settings \
         under lsp.{BINARY}.binary.path. A Zed launched from the Finder or Dock does not inherit \
         the $PATH your shell sets."
    )
}

impl zed::Extension for KomentExtension {
    fn new() -> Self {
        Self { found: None }
    }

    fn language_server_command(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<zed::Command> {
        let binary = match Self::configured_for_language_server(language_server_id, worktree) {
            Some(configured) => configured,
            None => worktree.which(BINARY).ok_or_else(nowhere_to_be_found)?,
        };
        self.found = Some(binary.clone());

        Ok(zed::Command {
            command: binary,
            args: vec!["lsp".to_string()],
            env: worktree.shell_env(),
        })
    }

    fn context_server_command(
        &mut self,
        context_server_id: &ContextServerId,
        project: &Project,
    ) -> Result<zed::Command> {
        let binary = Self::configured_for_context_server(context_server_id, project)
            .or_else(|| self.found.clone())
            .unwrap_or_else(|| BINARY.to_string());

        Ok(zed::Command {
            command: binary,
            args: vec!["mcp".to_string(), "--write".to_string()],
            env: Vec::new(),
        })
    }
}

zed::register_extension!(KomentExtension);

# Two-Tier Cascading Configuration

## Context
VoiceStudio/OmniVoice requires sensitive connection parameters (endpoint URL, Bearer authentication token), whereas projects have their own creative voice styling (speaker profile ID, language, character instruction, speed, and timeline slot durations). Storing tokens inside project manifests risks accidental disclosure when sharing or committing projects to Git.

## Decision
We decided to implement a two-tier configuration system. User-wide credentials and default connection settings are stored in `$XDG_CONFIG_HOME/vo-studio/config.json`. Individual projects define their voice styling, default slot duration, and act/frame structure in `vo-project.json`, inheriting global credentials unless explicitly overridden. Individual frames can optionally override voice parameters for dialogue.

## Considered Options
- **Purely Self-Contained Project Config**: Storing tokens inside each `vo-project.json` causes token duplication and leaks secrets into version control.
- **Environment Variables Only**: Reading only environment variables makes launching from diverse desktop/terminal environments fragile.

## Consequences
- Projects are safe to commit to version control without exposing API tokens.
- Running on a new machine requires setting up `$XDG_CONFIG_HOME/vo-studio/config.json` once.

# Hybrid Project Registry Storage Model

## Context
`vo-studio-tui` needs to run globally from any working directory while allowing users to create, organize, and version-control voiceover projects across the filesystem.

## Decision
We decided to adopt a hybrid storage model where `$XDG_DATA_HOME/vo-studio/registry.json` maintains a central index of registered projects, while projects themselves are self-contained directory trees with their own `vo-project.json`, `scripts/`, and `audio/` directories. Projects can be created in the default directory (`$XDG_DATA_HOME/vo-studio/projects/<project-slug>/`) or located in arbitrary directories (such as standalone git repositories) and registered into the hub.

## Considered Options
- **Pure Centralized Store**: Restricting all projects to `$XDG_DATA_HOME/vo-studio/projects/` prevents embedding projects in dedicated git repositories.
- **Pure Working Directory Mode**: Requiring `cd` into project folders before launching prevents running a global manager or project hub from anywhere.

## Consequences
- The Project Hub can discover, list, and switch between projects instantly from any terminal location.
- The application must handle missing or moved directories gracefully when reading the registry.

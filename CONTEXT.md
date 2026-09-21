# VoiceStudio TUI

A terminal interface for managing, synthesizing, and mastering structured voiceover productions.

## Language

**Project**:
A self-contained voiceover workspace containing configuration, narration scripts, and audio assets.
_Avoid_: Workspace, repository, job

**Project Registry**:
A central index stored in the user data directory that tracks known projects across the filesystem.
_Avoid_: Catalog, database, project list

**Project Hub**:
The primary launcher view listing known projects and facilitating selection, creation, and directory registration.
_Avoid_: Home screen, project picker, menu

**Global Config**:
User-wide configuration stored in the XDG config directory containing credentials and API connection defaults.
_Avoid_: User settings, master config

**Project Config**:
Project-specific settings stored in the project manifest specifying voice styling, timeline parameters, and structure.
_Avoid_: Project settings, project file

**Act**:
A top-level narrative section containing an ordered sequence of frames.
_Avoid_: Chapter, scene, sequence

**Frame**:
An individual timeline segment containing voiceover script text, rendering parameters, and duration constraints.
_Avoid_: Block, clip, cut

**Slot Duration**:
The allocated timeline window for a frame, padded with silence when positive, or unconstrained when zero.
_Avoid_: Frame length, timeline length, time limit

**Master**:
A fully assembled, continuous audio file representing an entire act or full production.
_Avoid_: Mixdown, final export, render

**Script**:
A plain-text document holding the narration script for an individual frame, stored decoupled on disk.
_Avoid_: Text file, narration doc, dialogue





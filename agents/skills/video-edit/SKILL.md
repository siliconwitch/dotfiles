---
name: video-edit
description: Edits OBS webcam and screen recordings, with b-roll and photos, into a 1440p landscape video and a separate portrait cut for Shorts and Reels, with framing, titles, portrait captions and technical animations. Transcribes with ElevenLabs Scribe, describes frames with Claude, and renders with ffmpeg. Use when Raj asks to edit, cut or caption a recording, to add b-roll, photos, titles or an animation, to change an edit, or to set up OBS for recording.
---

# Video edit

The program in this directory does the mechanical work: ingesting,
transcribing, describing frames, framing shots, and rendering. You write the
edit lists, render, check the result and deliver the videos. Aim to deliver
within minutes, because Raj uses this for spontaneous clips.

`edit-list.md` defines the edit list and how to cut. Read it before writing
one.

Run every step from this directory with an absolute folder path. The program
needs no C compiler, so `CGO_ENABLED=0` keeps `go run` working where none is
installed:

```sh
CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . <command> ~/videos/<folder> [argument...]
```

`go run .` with no command lists the commands.

## Recording setup

OBS records screen and webcam onto one canvas, so both share one clock. The
crop regions in `internal/project/sources.go` must match this layout:

- Canvas and output resolution: 4480x1440.
- Screen Capture (PipeWire): bounding box 2560x1440 at 0,0.
- Video Capture Device (webcam): bounding box 1920x1080 at 2560,0.
- Microphone on audio track 1.
- Recording format MKV, so a crash leaves a playable file. Recording path
  `~/videos/recordings`.
- Output mode Advanced, recording encoder QSV AV1, keyframe interval 1 s.
  Every segment of an edit seeks to the keyframe before it, and the default
  8 s interval doubles the time a render spends decoding.
- Tools > WebSocket Server Settings: server enabled, authentication off.

`Mod+Shift+6` toggles recording through `obs-cmd`. A notification shows OBS's
reply.

## Folders

Raj drops b-roll clips and photos from a phone or camera into
`~/videos/inbox/`. OBS writes captures to `~/videos/recordings/`. Raj clears
both, so `ingest` moves files out of them. It hard links a file from anywhere
else, such as b-roll from another project, which leaves that project intact.

Each video has one folder, `~/videos/<date>-<name>/`, which holds everything
needed to archive it once posted:

| Path | Written by | Holds |
|---|---|---|
| `brief.md` | you | what Raj wants from the edit |
| `sources/` | `ingest` | `recording-<n>.mkv`, `broll/`, `photos/` |
| `analysis/` | `analyse` | per source, `transcript.txt`, `transcript.json`, `frames/` and `frames.json`. `descriptions.txt` lists every source's details and frames |
| `edit/` | you | `landscape.json`, `portrait.json` |
| `animations/<id>/` | you, animator | `spec.md`, `index.html`, `animation.css`, `notes.md` |
| `build/` | `render`, `animate`, `sheet` | intermediate files, all regenerable. `build/cache/` holds rendered clips and segments for reuse |
| `out/` | `render` | `landscape.mp4` at 2560x1440, `portrait.mp4` at 1080x1920 |

`analyse` skips work it has already done, because transcription and frame
description cost money. It processes only new sources after a later ingest.
`build/` can be deleted before archiving. Keep `analysis/`.

## Steps

1. Create the folder, ingest the newest capture and everything in the inbox,
   and start analysing at once:

   ```sh
   mkdir ~/videos/<date>-<name>
   ls -t ~/videos/recordings/*.mkv | head -1
   ls ~/videos/inbox
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . ingest ~/videos/<folder> <capture> ~/videos/inbox/*
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . analyse ~/videos/<folder>
   ```

   `analyse` reads its keys from `pass` before anything else. When the store
   is locked, it stops within a second. Ask Raj straight away to run
   `pass show api/anthropic >/dev/null` in a terminal, then run `analyse`
   again once Raj confirms. Otherwise it takes about a minute for a few
   minutes of recording.
2. Write what Raj asked for to `brief.md`, including any facts for titles,
   such as names, roles, events, places and dates. When Raj gave no
   instructions, use the defaults in `edit-list.md` instead of asking.

3. Write `edit/landscape.json`, then `edit/portrait.json`, unless the brief
   asks for one format only. Run `check` on each until it passes with no
   warnings:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . check ~/videos/<folder> <format>
   ```

4. Write a spec and cutaway for each animation `edit-list.md` calls for, then
   spawn one animator per animation, all at once. Render a format that uses
   no animation while they work.

   ```text
   You are a sub-agent. Read ~/.config/agents/skills/video-edit/agents/animator.md and follow it.
   Project: ~/videos/<folder>. Animation: <id>.
   ```

   When its `notes.md` asks for more time or another cue, change the cutaway
   in each edit list that uses the animation and send the animator back to
   check its stills.
5. Render one format at a time, because each render uses every core. A render
   reuses every clip and segment that has not changed since the last one, so
   a fix renders again in seconds:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . render ~/videos/<folder> landscape
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . render ~/videos/<folder> portrait
   ```

6. Make contact sheets of each video and read every sheet listed:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . sheet ~/videos/<folder> <format>
   ```

   Fix and render again only for a clear fault: a subject cut off by the
   framing, a title, caption or corner cutaway over a face, a title with a
   wrong fact, a cutaway that does not match the words under it, a screen
   crop that cuts off what Raj points at, or a black or frozen frame. In portrait, also anything that matters in the bottom fifth or at
   the right edge, where Shorts and Reels lay their buttons.
7. Give Raj the paths of the videos.

## Changes from Raj

- **A cut, layout, framing, cutaway or title:** edit the edit list, run
  `check`, then `render`.
- **How an animation looks:** spawn an animator with the animation's id and
  Raj's words. It changes only the drawing, so the cues, and the cut around
  them, stay put. Then render.
- **When an animation's beats happen:** move its cues in the edit lists to the
  start of the words Raj wants them on. Spawn an animator only when a beat is
  added or removed.
- **Something Raj forgot to say:** Raj records a new clip. Ingest the newest
  capture, run `analyse`, and work the new recording, `recording-<n+1>.mkv`,
  into the edit lists.

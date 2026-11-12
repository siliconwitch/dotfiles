---
name: video-edit
description: Edits OBS webcam and screen recordings, talk or webinar exports, b-roll and photos into a 1440p landscape video, portrait cuts for Shorts and Reels, or both, with framing, titles, portrait captions and technical animations. Transcribes with ElevenLabs Scribe, describes frames with Claude, and renders with ffmpeg. Use when Raj asks to edit, cut or caption a recording, to make clips or Shorts, to add b-roll, photos, titles or an animation, to change an edit, or to archive a finished video.
---

# Video edit

The program in this directory does the mechanical work: ingesting,
transcribing, describing frames, framing shots, and rendering. You write the
edit lists, render, check the result and deliver the videos. Aim to deliver
within minutes, because Raj uses this for spontaneous clips.

`edit-list.md` defines the edit list and how to cut. Read it before writing
one.

Run every step with the project folder's absolute path. The program needs no C
compiler, so `CGO_ENABLED=0` keeps `go run` working where none is installed:

```sh
CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . <command> ~/videos/<project> [argument...]
```

The same command with no command or project lists the commands.

## Recordings

OBS records the screen and the webcam side by side on one canvas, so both share
one clock and one microphone track. `internal/project/sources.go` treats a file
the size of the canvas as a recording, and crops the screen and the webcam out
of it. The canvas size is in
`~/.config/obs-studio/basic/profiles/Untitled/basic.ini`, and each picture's
position on it is in `~/.config/obs-studio/basic/scenes/Untitled.json`. When
`ingest` files a capture under `broll/`, or a frame shows a crop cutting into
the wrong picture, compare those files with `sources.go` and tell Raj.

A file of another size is b-roll. When one holds a screen and a webcam side by
side, such as a webinar export, its segments name the regions to crop, as
`edit-list.md` describes.

## Folders

OBS writes captures to `~/videos/recordings/`. Raj copies clips and photos off
a phone or camera into `~/downloads/`, which also holds unrelated files.
`ingest` moves each file it is given out of either folder, and touches nothing
else. It symlinks a file from `~/vault`, because the vault already holds it,
and the archive keeps the symlink. It hard links a file from anywhere else,
such as b-roll from another project in `~/videos`, so this project keeps the
file when the other is archived and deleted. Each project's archive then holds
its own full copy.

Each video has one project folder, `~/videos/<project>/`. `<project>` is the
name its archive will have, with the time the edit starts added after the
date, such as `20261111-1430-sw-video-s2-air-quality-monitor`. Name it by the
project folder rules in the `vault-access` skill, with the category `video`.
The folder holds everything needed to rebuild the video from its archive:

```text
~/videos/<project>/
├── brief.md                      what Raj wants, plus facts for titles (you)
├── edit/                         (you)
│   ├── landscape.json
│   ├── portrait.json
│   └── portrait-<name>.json      a variant, such as one of several Shorts
├── animations/<id>/              (you, animator)
│   ├── spec.md
│   ├── index.html
│   ├── animation.css             copied from agents/
│   ├── animation.js              copied from agents/
│   └── notes.md                  timing requests, when any
├── out/                          (render)
│   ├── landscape.mp4             2560x1440
│   ├── portrait.mp4              1080x1920
│   └── portrait-<name>.mp4       1080x1920
├── .sources/                     (ingest)
│   ├── recording-<n>.mkv
│   ├── broll/<original name>
│   └── photos/<original name>    vault files are symlinks into ~/vault
├── .analysis/                    (analyse)
│   ├── descriptions.txt          every source's details and frame descriptions
│   └── <key>/                    such as recording-1 or broll-IMG_0412
│       ├── transcript.txt
│       ├── transcript.json
│       ├── frames.json
│       └── frames/<stream>-<nnnnnn>.jpg
└── .build/                       regenerable (render, animate, sheet)
    ├── cache/<hash>.<ext>        rendered clips, segments and captions
    ├── <format>/
    │   ├── segments.txt
    │   ├── sheet/<nnn>.jpg, <nnn>.txt
    │   └── sheet-<n>.jpg
    └── animations/<id>/<format>-cutaway<n>-<k>-<cue>.png
```

`analyse` skips work it has already done, because transcription and frame
description cost money. It processes only new sources after a later ingest.
Only the outputs and the files you write are visible, so a photo viewer shows
no sources, frames or render clips.

## Steps

1. Find the files for this video. Raj may add footage made weeks or months
   earlier, such as a clip from last week's event or photos of a board
   soldered months ago, so when a file arrived says nothing. List every
   capture, and every file under `~/downloads` with when and where it was
   made:

   ```sh
   ls -l ~/videos/recordings
   exiftool -q -r -T -Directory -FileName -CreateDate -GPSPosition ~/downloads
   ```

   Raj's request says what to look for. Files made on one day in one place
   are usually one shoot. Look at one frame of each capture and of each shoot,
   and take what fits the request. Ask Raj about anything that might belong
   but you cannot place.

   ```sh
   frame=$(mktemp --suffix=.jpg)
   ffmpeg -v error -y -i <file> -frames:v 1 "$frame"
   ```

2. Create the project folder. Its owner sets the prefix: `sw` when the video
   is about Silicon Witchery, its products or its events, and `rn` otherwise.
   Ask Raj when the owner is unclear:

   ```sh
   mkdir ~/videos/<project>
   ```

3. Ingest the files you picked, which moves them out of `~/downloads` and
   `~/videos/recordings`:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . ingest ~/videos/<project> <file>...
   ```

4. Look for archived footage to work in. Read the `vault-access` skill first.
   Look only in `~/vault/company-media/` for an `sw` video, or only in
   `~/vault/personal-media/` for an `rn` one. Each folder there is one archived
   project, such as `20251007-sw-video-s2-air-quality-monitor`. In a project
   this skill archived, a `.rclonelink` file holds the path of a clip in
   another vault folder. Use that path instead.

   Take a clip when Raj mentions an older video, or when it shows what Raj
   talks about better than the new footage. Ingest only clips you will use,
   because `analyse` downloads each one through the mount and pays to
   transcribe and describe it:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . ingest ~/videos/<project> ~/vault/<company-media|personal-media>/<archived project>/<file>...
   ```

5. Analyse every source:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . analyse ~/videos/<project>
   ```

   `analyse` reads its keys from `pass` before anything else. When the store
   is locked, it stops within a second. Ask Raj straight away to run
   `pass show api/anthropic >/dev/null` in a terminal, then run `analyse`
   again once Raj confirms. Otherwise it takes about a minute for a few
   minutes of recording.

   Then read every `transcript.txt`. When Raj mentions a picture, a clip or an
   older video you have not ingested, find it in `~/downloads` as in step 1 or
   in the vault as in step 4. Ingest it and run `analyse` again.

6. Write what Raj asked for to `brief.md`, including any facts for titles,
   such as names, roles, events, places and dates. When Raj gave no
   instructions, use the defaults in `edit-list.md` instead of asking.

7. Write `edit/landscape.json`, then `edit/portrait.json`, unless the brief
   asks for one format only or for variants. Run `check` on each until it
   passes with no warnings:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . check ~/videos/<project> <format>
   ```

8. Write a spec and cutaway for each animation `edit-list.md` calls for, then
   spawn one animator per animation, all at once. `<id>` is the animation's
   folder name in `animations/`, such as `data-flow`. Render a format that
   uses no animation while they work.

   ```text
   You are a sub-agent. Read ~/.config/agents/skills/video-edit/agents/animator.md and follow it.
   Project: ~/videos/<project>. Animation: <id>.
   ```

   When its `notes.md` asks for more time or another cue, change the cutaway
   in each edit list that uses the animation and send the animator back to
   check its stills.

9. Render one edit list at a time, because each render uses every core. Name
   a variant by its edit list, such as `portrait-five-minute-fix`. A render
   reuses every clip and segment that has not changed since the last one, so
   a fix renders again in seconds:

   ```sh
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . render ~/videos/<project> landscape
   CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . render ~/videos/<project> portrait
   ```

10. Make contact sheets of each video and read every sheet listed:

    ```sh
    CGO_ENABLED=0 go -C ~/.config/agents/skills/video-edit run . sheet ~/videos/<project> <format>
    ```

    Fix and render again only for a clear fault: a subject cut off by the
    framing, a title, caption or corner cutaway over a face, a title with a
    wrong fact, a cutaway that does not match the words under it, a screen
    crop that cuts off what Raj points at, or a black or frozen frame. In
    portrait, also anything that matters in the bottom fifth or at the right
    edge, where Shorts and Reels lay their buttons.

11. Give Raj the paths of the videos.

## Changes from Raj

- **A cut, layout, framing, cutaway or title:** edit the edit list, run
  `check`, then `render`.
- **How an animation looks:** spawn an animator with the animation's id and
  Raj's words. It changes only the drawing, so the cues, and the cut around
  them, stay put. Then render.
- **When an animation's beats happen:** move its cues in the edit lists to the
  start of the words Raj wants them on. Spawn an animator only when a beat is
  added or removed.
- **More footage:** Raj records a new clip or adds files to `~/downloads`.
  Repeat steps 1, 3, 4 and 5 for `~/videos/<project>`, then work the new
  sources into the edit lists.

## Archive

When Raj says the video is finished, copy the project folder to the vault,
following the `vault-access` skill, but use these steps in place of its
Archive steps. Leave out `.build/`, which `render` makes again.
`<archived project>` is the project folder's name without the time, such as
`20261111-sw-video-s2-air-quality-monitor`. `--links` stores each vault
symlink as a small `.rclonelink` file, so a project copied back with
`--links` finds its vault clips again without a second copy.

1. Copy the project:

   ```sh
   rclone copy --links --exclude '/.build/**' ~/videos/<project> vault:<company-media|personal-media>/<archived project>
   ```

2. Check every file arrived intact, and report the result:

   ```sh
   rclone cryptcheck --links --exclude '/.build/**' ~/videos/<project> vault:<company-media|personal-media>/<archived project>
   ```

3. When `cryptcheck` reports no differences, delete `~/videos/<project>`. Raj
   does not need to confirm.

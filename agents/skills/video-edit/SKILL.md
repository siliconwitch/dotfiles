---
name: video-edit
description: Edits videos for YouTube (landscape) and social media posts (portrait) from footage already on this machine. Use when Raj asks for an edit, or for social media posts that need video.
---

# Video edit

The workflow is strict. You make creative choices only in step 4, the timeline, and step 6, the check. Every
other step is a command or a sub-agent, run exactly as written.

- **Footage** is a video or image file.
- **A command** runs the Go app in this skill's directory. Give every path in full, because `go -C` runs the
  app from its own directory.
- **A sub-agent** step runs in a sub-agent. Tell it that it is a sub-agent, give it the step's instructions
  and what they need, and take back only its result. Its working has no later use and would fill your
  context.

A project lives in `~/videos/<project>/`:

| Path | Holds |
|---|---|
| `pre-analysis.txt` | a summary of each candidate file, from step 2 |
| `.footage/` | the footage, moved or linked in step 2 |
| `analysis/transcripts/<file>.txt` | each file's speech, one phrase per line |
| `analysis/frames/<file>.txt` | each file's frames, described |
| `analysis/.data/` | the app's own data. Never read or edit it |
| `timelines/<name>.txt` | one timeline per video, from step 4 |
| `animations/<id>/` | each animation: your `spec.md`, and the animator's `scene.json` and `notes.md`, from step 4 |
| `.build/<name>-<n>/` | each quick render, from step 5 |
| `.build/animations/<id>/` | the animator's stills, from step 4 |
| `out/<name>.mp4` | each finished video, from step 7 |

## 1. Prepare

Split every new OBS recording into its screen and webcam footage:

```sh
CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . prepare ~/videos/recordings
```

Then name the project with the time now and a short topic, such as `20261112-1656-sw-video-air-quality-monitor`,
and create its footage folder:

```sh
mkdir -p ~/videos/<project>/.footage
```

Tell Raj the project's full path, and use it in every later step.

## 2. Find the footage (sub-agent)

Give the sub-agent Raj's request, the project's full path, and these steps. It returns the full path of every
file in `.footage/`, one per line, and nothing else.

1. List every file that could be footage, broadly, because nothing shows what a file holds yet. Search each
   place below recursively, because footage often sits in subfolders, at any depth:
   - Every file in `~/videos/recordings` whose name does not start with a dot.
   - Every file under `~/downloads`, except ones clearly unrelated, such as documents, installers and
     archives.
   - In the vault, read the `vault-access` skill first. Find the project folders in `~/vault/company-media/`
     whose names might relate to the request, or in `~/vault/personal-media/` for Raj's personal videos, then
     list their files.
2. Summarise them all:

   ```sh
   CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . pre-analyse -o ~/videos/<project>/pre-analysis.txt <file>...
   ```

3. Shortlist the files the request needs from `pre-analysis.txt`. Keep a recording's screen and webcam files
   together.
4. Move each shortlisted file from `~/videos/recordings` or `~/downloads` into `~/videos/<project>/.footage/`
   with `mv -n`, and link each vault file there with `ln -s`. Never overwrite a file. When two files share a
   name, prefix the second with its folder's name, such as `trip-IMG_0412.MOV`.

## 3. Analyse

Analyse every file in `.footage/`:

```sh
CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . analyse -o ~/videos/<project>/analysis ~/videos/<project>/.footage/*
```

It resumes work already done, so when it fails partway, run it again with the same files. When `pass` is
locked, ask Raj to unlock it as the error says, then run it again.

## 4. Build the timeline

Read every file in `analysis/transcripts/` and `analysis/frames/` before writing anything, and choose every
shot from what the frames show. Then write `timelines/<name>.txt` as `timeline.md` describes, following the
editing rules below.

A full video, a landscape video and a YouTube video are the same thing. Name its timeline `landscape`. Name each
Short `portrait-<topic>`, such as `portrait-five-minute-fix`. When Raj asks for clips or Shorts, write only
portrait timelines, unless he also asks for a full video.

### Animations (sub-agent)

For each animation the editing rules call for:

1. Write `animations/<id>/spec.md`, with a short kebab-case id such as `data-flow`. An animator draws it from
   the spec alone, so write it for someone who has not heard the recording. Name the place it fills in each
   timeline, and whether Raj sits in the corner. Describe each part, the links between them, and what moves
   along the links. List the cues in order, each a short kebab-case name with the words that trigger it. Keep
   to about eight parts, or one for a corner animation.
2. Add its `b` line to each timeline that uses it, with a word for every cue, as `timeline.md` describes.
3. Start one sub-agent per animation, all at once, with this prompt:

   ```text
   You are a sub-agent. Read ~/.claude/skills/video-edit/animator.md and follow it.
   Project: ~/videos/<project>. Animation: <id>. Timelines: <name>, <name>.
   ```

Write the rest of the timeline while they work. `quick-render` fails until every animation has its
`scene.json`. When the check or Raj finds a fault in how an animation looks, send its animator back with only
that fault. When a beat lands on the wrong word, move its word in `cues` yourself. When an animator reports a
timing request, give the beat more time, such as by moving the next cue's word, or ask Raj.

## 5. Quick render

```sh
CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . quick-render ~/videos/<project> <name>
```

It writes a build to `.build/<name>-<n>/`:

| File | Holds |
|---|---|
| `preview.mp4` | the cut at low resolution |
| `script.txt` | each layer a clip's time, timeline line and words |
| `heard.txt` | a fresh transcription of the preview |
| `cuts.txt` | every cut: its time, the lines either side, and the words either side. Then every title and cue |
| `sheet-NN.jpg` | the last frame before and the first frame after every cut, and frames before and 1 s after every title and cue, labelled as in `cuts.txt` |

## 6. Check

Run the review in a sub-agent. Give it the build's full path, the timeline, Raj's request and the editing
rules below. It reads every file in the build, and returns each problem with its cut number, its timeline line
and what it saw. Ask it to look for:

- words in `heard.txt` that differ from `script.txt`, such as a clipped word, a leaked word or a repeat
- a cut that breaks a sentence or lands mid-thought
- a head cut off, a face out of frame, or a shot of a hand, a wall or an empty desk
- a flash: a clip of a frame or two between two others
- the wrong picture for the words, such as b-roll that shows something else
- a title or caption over a face, or a title that states something Raj did not
- an animation beat that lands before or long after the words about it, or a part that is hard to read

Fix what it reports by going back to step 4. When the build is right, give Raj the full path of
`preview.mp4` and wait for him. Each change he asks for goes back to step 4.

## 7. Render

Once Raj approves the preview:

```sh
CGO_ENABLED=0 go -C ~/.claude/skills/video-edit run . render ~/videos/<project> <name>
```

It renders only a timeline that has had a quick render since it last changed. Give Raj the full path of
`out/<name>.mp4`. A video for social media follows the `marketing` skill.

## Editing rules

### Story

- Tell one story, in the order Raj asks for or the order he told it.
- Keep the last take of a repeated line. Raj restarts a line when the take before was wrong.
- Cut filler words, false starts such as "add it to--", and long silences.
- Follow a note Raj speaks to the editor, such as "to the editor, cut the last minute", and cut the note
  itself. Ask Raj when a note is unclear.
- End every passage on a complete thought, never mid-sentence, such as on "connected to their,".
- Read `script.txt` as one script. It must read as Raj telling one story, each clip following from the last.
- Never invent a fact.

### Transcripts

- A `-- 84.2 s without speech --` line is a gap. Read the frames to see what happens in it, such as Raj
  leaving the camera to reset a demo.
- A note such as `("Blah." lasts 5.3 s)` marks noise heard as a word. Leave it out of the cut.

### Titles

- Add a title where the viewer needs a name, such as Raj's first shot, a guest's first shot, or b-roll of an
  event while Raj talks over it.
- Title Raj's first shot `Raj Nakarja · Founder · Silicon Witchery`, unless Raj asks for another or for none.
- Title b-roll of an event with the event and its date, such as `Onomondo Demo Day - 20 May 2026`.
- Take every fact from Raj, his request, a vault folder's name, or the start date in a frames file. Never
  invent a name, role, place or date. When a fact is missing, leave the title out.

### Animations

- Add an animation where Raj explains something technical that a drawing shows better than words, such as how
  data flows between parts, what is inside a device, or how a protocol works.
- Add none while Raj shows something physical, such as a board held up to the camera, or works on the screen.
  There the picture is the point.
- In landscape, fill the picture with the animation, with Raj in the corner when his face keeps the
  explanation personal. Or keep Raj or his work in the picture, with a small animation of one thing in the
  corner.
- In portrait, put the animation on top with Raj below, or fill the picture with it.
- Cover 8 to 30 s of the explanation, and give each beat about 1 s before the next.

### Pictures

- Show the camera Raj faces. Each frame's framing says where he looks.
- Change between the camera and the screen only when what matters changes, such as from talking to a demo.
  Switching back and forth is jarring.
- Show the screen with the webcam in the corner when Raj talks over what the screen shows. Crop the screen to
  the window he works in with a `frame` box, and keep one box for the whole passage.
- Put b-roll and photos on the words about them. Show a photo for 2 to 4 s.
- Short clips are fine. Raj minds a cut that lands inside a word, not a short one.
- Fix a framing fault in the sheets, such as a chin cut off or a head out of frame, with `frame=full` or a
  `frame` box.

### Portrait

- Keep to the length Raj gives. Otherwise aim for 30 to 60 s, and let a story that needs it run to 2 minutes.
- Open on the most striking line or picture within the first 2 s. It need not come first in the recording.
- Keep one idea. Drop asides and setup the hook does not need.
- End on the payoff: the result, the number or the lesson the hook promises. Never end on setup, or on a
  problem without its answer.
- Prefer the webcam, or the screen above the webcam. Show the screen alone only with a `frame` box around the
  part worth reading, because code at phone width is unreadable.

## Changes later

When Raj asks to change a finished edit, start at step 4 with the existing project.

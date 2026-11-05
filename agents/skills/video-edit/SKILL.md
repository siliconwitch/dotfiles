---
name: video-edit
description: Edits a webcam and screen recording from OBS into a captioned social media video. Transcribes the audio with ElevenLabs Scribe, describes frames with Claude, plans the cuts and layouts from a brief, and renders with ffmpeg. Use when Raj asks to edit, cut or caption a recording, to change an edit, or to set up OBS for recording.
---

# Video edit

The program in this directory does the transcription, frame description and
rendering. You do the editing: you read the transcript and frame descriptions,
decide the cuts and layouts, and write `edl.json`.

Run every step from this directory with an absolute folder path:

```sh
go -C ~/.config/agents/skills/video-edit run . <step> ~/videos/<folder>
```

## Recording setup

OBS records screen and webcam onto one canvas, so both share one clock. The
crop regions in `internal/canvas/canvas.go` must match this layout:

- Canvas and output resolution: 4480x1440.
- Screen Capture (PipeWire): bounding box 2560x1440 at 0,0.
- Video Capture Device (webcam): bounding box 1920x1080 at 2560,0.
- Microphone on audio track 1.
- Recording format MKV, so a crash leaves a playable file. Recording path
  `~/videos/recordings`.
- Tools > WebSocket Server Settings: server enabled, authentication off.

`Mod+Shift+6` toggles recording through `obs-cmd`. A notification shows OBS's
reply.

## Folder

One folder per video, `~/videos/<date>-<name>/`:

| File | Written by | Holds |
|---|---|---|
| `source.mkv` | you | a hard link to the OBS recording |
| `brief.md` | you | what Raj wants from the edit |
| `transcript.json` | `transcribe` | every word and sound with start and end in seconds |
| `transcript.txt` | `transcribe` | one phrase per line, `start-end text` |
| `frames/` | `frames` | screen and webcam JPEGs and `index.json` |
| `descriptions.txt` | `describe` | one line per frame, `time stream file: description` |
| `edl.json` | you | the edit |
| `captions.ass`, `out/edit.mp4` | `render` | the captions and the finished video |

`transcribe`, `frames` and `describe` skip themselves when their output
exists, because each costs time or money. Delete the output only when Raj asks
to redo that step. `render` always runs.

## Steps

1. Create the folder and link the recording into it. A hard link leaves the
   original in place and uses no space:

   ```sh
   mkdir ~/videos/<date>-<name>
   ln ~/videos/recordings/<recording>.mkv ~/videos/<date>-<name>/source.mkv
   ```

2. Write Raj's requirements to `brief.md`. Ask for them if none were given.
3. Transcribe:

   ```sh
   go -C ~/.config/agents/skills/video-edit run . transcribe ~/videos/<folder>
   ```

4. Extract frames:

   ```sh
   go -C ~/.config/agents/skills/video-edit run . frames ~/videos/<folder>
   ```

5. Describe the frames. This uses `claude-sonnet-5-5` unless you add a model
   ID after the folder. The step waits for the batch, often under an hour, and
   resumes the same batch if interrupted:

   ```sh
   go -C ~/.config/agents/skills/video-edit run . describe ~/videos/<folder>
   ```

6. Read `brief.md`, `transcript.txt` and `descriptions.txt`, and write
   `edl.json`. Open a frame with Read when its description leaves a decision
   unclear.
7. Render:

   ```sh
   go -C ~/.config/agents/skills/video-edit run . render ~/videos/<folder>
   ```

8. Check the edit by extracting a frame from each segment of `out/edit.mp4`
   and reading it. Then give Raj the path.

For a change to an edit, edit `edl.json` and run `render` again.

## Locked pass

`transcribe` reads `api/elevenlabs` and `describe` reads `api/anthropic` from
`pass`. When the store is locked, the step stops at once and says so. Tell Raj
to run `pass show api/anthropic >/dev/null` in a terminal, wait for Raj to
confirm, then run the step again.

## The edit

`edl.json` lists segments in playback order. Times are seconds in
`source.mkv`:

```json
{
  "segments": [
    {"start": 2.10, "end": 14.80, "layout": "webcam"},
    {"start": 15.30, "end": 61.00, "layout": "screen+webcam"},
    {"start": 64.20, "end": 90.50, "layout": "screen"}
  ]
}
```

- `webcam`: the webcam full frame. Use it when Raj talks to the camera or
  shows hardware.
- `screen`: the screen full frame. Use it when the screen is the whole point,
  such as code too small to read with an inset over it.
- `screen+webcam`: the screen with the webcam in the bottom right. Use it when
  Raj talks over what the screen shows.

Cut on word boundaries from `transcript.json`. Start a segment about 0.1 s
before its first word and end it about 0.2 s after its last, so no word is
clipped. Cut filler words, false starts, long silences, and every take but
the last of a repeated line. The brief overrides these defaults.

The output is 1920x1080 at 30 fps, with captions burned in from the words each
segment keeps.

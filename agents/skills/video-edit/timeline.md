# Timeline

A timeline, `timelines/<name>.txt`, is one video. It has two layers. Layer a plays in order and sets the
video's length. Layer b covers part of layer a, such as a phone angle, b-roll, a photo, an animation, or the
webcam in the corner of the screen. A `title` line labels the line above it.

```
format landscape

# The opening, to the camera
a 20261112-155546-webcam-capture.mkv p1-p14
title Raj Nakarja · Founder · Silicon Witchery
a 20261112-155546-webcam-capture.mkv p17-p22
b IMG_4698.MOV sync over=p19-p22

# The demo, with the webcam in the corner
a 20261112-155546-screen-capture.mkv p23-p40 frame=0,0,1800,1440
b 20261112-155546-webcam-capture.mkv sync place=corner

# How a reading reaches the phone, drawn, with Raj in the corner
a 20261112-155546-webcam-capture.mkv p41-p46 place=corner
b animations/data-flow 0 cues=sensor:p41.3,server:p43,phone:p45.2

# The result, with b-roll of the board
a 20261112-155546-webcam-capture.mkv p47-p52
b 20251007-sw-footage-s2-air-quality-monitor-04.mov 12.0 over=p48-p50
```

The first line is `format landscape` or `format portrait`. A line starting with `#` is a note. A `title` line
is described below. Every other line is `<layer> <file> <span> [option=value...]`, where `<file>` is a file
name in `.footage/`, or `animations/<id>` for an animation.

## Layer a

Each `a` line plays its span of the file, after the line before it. Its span is one of:

- Phrases from the file's transcript: `p12-p18` plays from the first word of p12 to the last word of p18.
  `p12.3-p18.2` starts at the third word of p12 and ends on the second word of p18. `p12` is one phrase, and
  `p12.3` is one word.
- Seconds of a file without a transcript, such as `3.00-9.50`, or how long to show a photo, such as `0-3`.

The app places each cut just before the first word and just after the last, never more than halfway into the
silence beside it, so no word is clipped and none of the next leaks in. It shortens every pause of over 1 s
inside a span to the usual gap between two words.

## Layer b

Each `b` line covers the `a` line above it, or part of it. Its span is one of:

- `sync`: the same moment from another file, such as the phone angle of what the webcam shows, or the webcam
  beside the screen. Every file was placed on one clock in step 3, so the moment matches at every cut.
- The second of the file to start at, such as `12.0`, for b-roll. Use `0` for a photo.

`over` names the part of the `a` line above to cover, as phrases or seconds of that line's file, such as
`over=p44-p46`. It may run on into later `a` lines of the same file. Without it, the `b` line covers the whole
`a` line above. Layer b holds one clip at a time.

## Animations

An animation is a `b` line whose file is `animations/<id>` and whose span is `0`. It is drawn for as long as
it covers, from `animations/<id>/scene.json`. It takes only `place`, `over` and `cues`:

- `place` is `full` or `corner` in landscape, and `full` or `top` in portrait.
- `cues` gives each cue in the scene a word of the `a` line's file, such as
  `cues=sensor:p41.3,server:p43,phone:p45.2`. `p43` is the first word of p43. Each beat starts with its word, so a
  recut keeps every beat on its word. Every cue in the scene needs a word, and every word must fall in the part
  of the cut the animation covers.

Start an animation's `over` on the words of its first cue. Before its first cue, an animation is empty.

To show Raj over a full animation, put the `a` line in the corner with `place=corner`. Layer a fills the corner
only while a `b` line fills the picture behind it, and the animation keeps its parts clear of the corner.

## Titles

`title <text>` puts the text on the picture of the line above it, `a` or `b`, such as
`title Onomondo Demo Day - 20 May 2026`. It fades in 0.3 s after that line starts, and shows for 6 s or until the line ends. It sits
bottom left in landscape, and above the captions in portrait, where it fits up to 40 characters. A line takes
one title.

## Captions

A portrait video shows captions of what Raj says, from the transcript, up to three words at a time. A landscape
video has none, because YouTube captions it.

## Options

| Option | Values | Default |
|---|---|---|
| `place` | `full`, `top`, `bottom` or `corner` | `full` |
| `frame` | `auto`, `full`, or a box such as `120,80,1800,1200` | `auto` |
| `speed` | 0.25 to 100 | 1 |
| `volume` | 0 to 4 | 1 on layer a, 0 on layer b |
| `pauses` | `keep`, on layer a: leave pauses of over 1 s as they are | shortened |

`place` puts the picture in a panel of the video:

| Place | Landscape | Portrait |
|---|---|---|
| `full` | the whole picture | the whole picture |
| `top`, `bottom` | the top or bottom half | the top or bottom half, such as the screen above the webcam |
| `corner` | a small 16:9 picture, bottom right | a small 16:9 picture, top right |

The corner always shows above the rest of the picture.

`frame` decides what of the file fills the panel:

- `auto` keeps every subject the frames file lists for the clip in view, and zooms at most 1.2 times. The
  clips in a row of one file, place and frame share one crop, so a jump cut never moves the camera. When the
  subjects range wider than the panel, as in a portrait crop of a person who moves around, it centres on where
  they usually are. When under half the frames list a subject, it shows the whole picture.
- `full` shows the whole picture, over a blurred copy of itself where its shape differs from the panel's.
- A box, `left,top,right,bottom` in pixels of the file, as the frames file gives subjects, shows that region,
  widened to the panel's shape. Measure it from the subjects in the frames file.

`speed` changes the pace, such as `speed=4` for a time-lapse of a build. Sound keeps its pitch.

`volume` sets the clip's sound. Layer b is silent by default, so a `sync` angle never doubles Raj's voice. Give
b-roll a volume, such as `volume=0.15`, only when its own sound adds something.

## Errors

`quick-render` checks the whole timeline before it renders, and names the line of every fault, such as a phrase
the transcript lacks, b-roll too short for its span, or a cue with no word.

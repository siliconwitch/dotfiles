---
name: vault-access
description: Rules for reading and archiving files in the encrypted rclone vault at ~/vault, including folder naming. Use before reading, listing or writing anything in ~/vault or vault:, when Raj asks to archive photos, videos, music, backups or other large files, or when another skill archives files to the vault.
---

# Vault access

The vault is the encrypted rclone remote `vault:`, mounted at `~/vault`.

## Vault folders

Use only the vault folders that belong to the repository you started in:

| Repository   | Vault folders                           | Prefix |
| ------------ | --------------------------------------- | ------ |
| `~/personal` | `personal-media`                        | `rn`   |
| `~/company`  | `company-media`, `company-large-archive` | `sw`   |

An agent started anywhere else uses the folders that match the owner of the files, such as `personal-media` for an `rn` video. Ask Raj when the owner is unclear.

## Access rules

Check that the vault is mounted before using it:

```sh
mountpoint -q ~/vault
```

If the check fails, or an rclone command fails with a `password-command` error, ask Raj to unlock `pass`, then check again. Never run `pass` or touch `vault.service`.

- Read only your vault folders, through `~/vault`. Never create, change, rename or delete anything in `~/vault`.
- Write with `rclone copy` to your vault folders under `vault:` only. Never run `sync`, `move`, `delete`, `purge` or `rmdirs`.
- Never use `storagebox:`. It is unencrypted.

## Project folders

Every folder directly under a vault folder is one project, such as one event, one trip or one video. Name each new project folder by the File naming rules in the owning repository's `CLAUDE.md`, using the categories already beside it. Date it with the day it was created, or with its earliest capture date when it holds photos or videos. Every agent must follow this folder naming.

## Default file layout

These rules are defaults for an agent started in `~/personal` or `~/company` when Raj asks it to archive files without saying how to name them. Other agents, and skills such as `video-edit`, may lay out and name the files inside a project folder their own way.

- Put every file directly in the project folder, with no subfolders.
- Name each photo or video after its folder plus an index in capture order, zero-padded to the same width across the folder, with a lowercase extension, for example `<folder>/<folder>-001.heic`. A Live Photo's image and video share one index.
- Name every other file by the File naming rules in the repository's `CLAUDE.md`.
- Put scans and reference images that are really documents in the repository's `archive/` instead.

## Archive

1. Put the files in the repository's `working/vault/<folder>/`, named as they should appear in the vault, where `<folder>` is the project folder's name, such as `20240421-rn-music-dark-dnb`.
2. Copy the folder, where `<area>` is one of your vault folders, such as `personal-media`:

   ```sh
   rclone copy working/vault/<folder> vault:<area>/<folder>
   ```

3. Check every file arrived intact, and report the result:

   ```sh
   rclone cryptcheck working/vault/<folder> vault:<area>/<folder>
   ```

4. Ask Raj before deleting `working/vault/<folder>/`.

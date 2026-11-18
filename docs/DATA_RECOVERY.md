# Recovering an interrupted OpenCMMC assessment

OpenCMMC stores assessment data on the local machine. It is not an encrypted
backup system, and it does not upload workspaces to BlockSiFr.

## Interrupted initial workspace creation

The app writes `requirements.json` (Level 1) or `objectives.json` (Level 2)
before writing `state.json`. If the process stops in this narrow interval,
a later startup may find a valid assessment without its state file.

OpenCMMC now infers the missing state **in memory** when exactly one validated
workspace exists. The existing data and evidence are left unchanged. The
browser can display and export the recovered assessment, and a new
`/create` attempt cannot overwrite it.

If both workspace files exist, a file is corrupt, or a file is unsafe (such
as a symlink), OpenCMMC fails closed instead of guessing. Do not delete or
edit the files to make the error disappear. Preserve the directory and seek
qualified support through the project's private support channel.

## Safe manual backup

1. Stop OpenCMMC using Ctrl+C. Close tabs that contain the local workbench.
2. On an approved encrypted device, copy the **entire** local
   `private-assessment` directory, including the `evidence` subdirectory,
   or the exact directory supplied by `--data-dir`.
3. Store the copy in an access-controlled backup location authorized for its
   data classification. Preserve file names, relative paths, and timestamps.
4. Confirm that the original and backup have the same file count and
   SHA-256 digests. Use approved procedures on the relevant operating system.
5. Test the backup with a **separate** `--data-dir`, preferably using
   synthetic evidence, before trusting any recovery procedure.
6. Never send assessment files, access codes, screenshots containing the
   private URL, or potentially controlled information in a public GitHub issue.

Do not copy a live workspace while forms are changing. Files may be
individually atomically written without constituting a transactionally
consistent multi-file snapshot.

## Limitations

Recovery of an interrupted initial creation is not full crash recovery,
versioned backup, rollback, ransomware resilience, or enterprise disaster
recovery. OpenCMMC does **not** encrypt evidence at the application layer.
Use approved operating-system encryption and retention practices. An
organization must separately approve any use involving CUI.

For general install and browser troubleshooting, see
[INSTALL_V1.md](INSTALL_V1.md). For security boundaries, see
[SECURITY.md](../SECURITY.md).

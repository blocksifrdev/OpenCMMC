# Offline assessment backups, verification and restoration

**Availability:** This feature is merged in the latest OpenCMMC source and signed
CI test artifacts after November 18, 2026. It is **not** included in the
previously published `v1.0.0-rc.11` portable archives. Verify the installed
build before attempting these commands; do not assume a README on `main`
describes an older public download.


OpenCMMC v1 workspace data is stored on the local machine. Backups are
complete snapshots of the workspace, its state file, and **only referenced**
evidence whose SHA-256 hashes match the workspace.

**Security warning:** Backup ZIP files are **NOT ENCRYPTED** by OpenCMMC.
Store backup and restored files only on approved encrypted storage. Backups
are not a CUI system authorization or evidence of effective compliance.

## Make a verifiable backup

1. Stop OpenCMMC (Ctrl+C) and close its local browser tabs. Backing up while
   writes occur may produce a snapshot from inconsistent moments.
2. Choose an approved encrypted drive and an output filename that does not
   exist. Never write backups inside the live assessment directory.
3. Run:

```sh
./opencmmc cmmc backup \
  --data-dir ./private-assessment \
  --archive /approved/encrypted-backups/assessment.zip

./opencmmc cmmc verify-backup \
  --archive /approved/encrypted-backups/assessment.zip
```

The backup fails closed for unsupported workspace levels, ambiguous workspace
types, missing or mismatched referenced evidence, symlinks, unsafe file types,
and oversize data. Archives contain no browser access token or telemetry.
The manifest records each member's SHA-256 digest and byte size.

A **matching integrity digest does not establish who made the archive**.
Keep backups under an organization-approved access and identity boundary.
The tool does not encrypt backups, digitally sign them or provide automatic
scheduling or disaster recovery.

## Restore without overwriting an existing assessment

```sh
./opencmmc cmmc restore-backup \
  --archive /approved/encrypted-backups/assessment.zip \
  --out-dir ./restored-assessment
```

The restore destination **must not already exist**. The archive is fully
validated before any files are written, including every member digest, the
assessment baseline, and referenced evidence. A restored workspace is
verified again before a successful completion is reported.

To view the restored assessment on the same computer, launch:

```sh
./opencmmc cmmc ui --data-dir ./restored-assessment --port 8765
```

Keep the new private session URL secret. Preserve the original assessment.
Do not replace production assessment data without your organization's
written recovery authorization.

## Scope, constraints and evidence

The feature accepts an archive of at most 512 MiB uncompressed total with
individual files capped at 50 MiB and a maximum of 4096 members. The
browser UI itself limits new attachments to 10 MiB each. Data is processed
locally with no remote transfer. This feature does not implement continuous
backups, automated restore testing, backup encryption, remote storage,
multi-user change control or key rotation.

Tests include both workspace levels, integrity tampering, symlink rejection,
ZIP traversal rejection, conflicting assessment state, refusal to overwrite,
and byte-identical restore verification.

Security and program boundaries: [SECURITY.md](../SECURITY.md),
[RELEASE_GATES_2026-11-11.md](RELEASE_GATES_2026-11-11.md).

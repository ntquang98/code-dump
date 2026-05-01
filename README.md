# codedump

`codedump` walks a source tree and produces two text files you can paste into an AI chat:

- `src_tree_<name>.txt`: a project structure tree (gitignore-aware)
- `<name>_dump.txt`: concatenated file contents (gitignore-aware)

## Install / build

### From source (recommended)

```bash
go build -o codedump ./app/cli
```

On Windows you may prefer:

```powershell
go build -o codedump.exe ./app/cli
```

## Usage

```bash
codedump --input <path-to-project> --name <name>
```

Example:

```bash
codedump --input . --name myproj
```

This writes:

- `src_tree_myproj.txt`
- `myproj_dump.txt`

Both files are created in your **current working directory** (the directory you run `codedump` from).

## Output format

### Tree (`src_tree_<name>.txt`)

ASCII tree with `/` suffix on directories.

### Dump (`<name>_dump.txt`)

For each included file:

1. Write the file path (relative to `--input`) plus `\n`
2. Write the raw file content
3. If the file does not end with a newline, `codedump` appends one

## Ignore rules

- Respects `.gitignore` rules (including nested `.gitignore` files).
- Always skips `.git/`.

## Safety / limitations

- **Symlinks are skipped** (not followed) to avoid cycles and unexpected traversal.
- **Binary-ish files are skipped from the dump** if they contain a NUL byte. (They may still appear in the tree if not ignored.)
- If any file read fails, the command exits non-zero.

## Notes

- `--name` must be a simple token (no path separators). It is used only to build output filenames.


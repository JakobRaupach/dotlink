# dotlink

A small Go CLI that symlinks the contents of a source directory (e.g. your dotfiles repo) into a target directory, with gitignore-style ignore patterns. Inspired by GNU Stow.

## Install

```bash
go install github.com/JakobRaupach/dotlink@latest
```

## Usage

```bash
dotlink [flags] <src>
```

Flags must come before `<src>`.

| Flag | Description | Default |
|---|---|---|
| `-dest` | Directory to create the symlinks in | parent of `<src>` |
| `-i` | Comma-separated ignore patterns | none |
| `-v` | Verbose output | `false` |
| `-q` | Only output errors | `false` |
| `-D` | Delete symlinks linked to src | `false` |
| `-R` | Reloads symlinks |  `false` |
| `-n` | Simulates the result | `false` |

### Example

```bash
# link everything in ~/dotfiles into ~, skipping swap files and the scripts dir
dotlink -dest ~ -i "*.swp,scripts/" ~/dotfiles
```

## How it works

- Walks `<src>` and creates a symlink in `-dest` for each entry, mirroring the directory layout.
- If a directory doesn't exist in the destination, the whole directory is symlinked. If it already exists, dotlink descends into it and links its contents individually.
- Existing files in the destination are left untouched.

## Ignoring files

Patterns use gitignore syntax. They come from, in order of increasing priority:

1. Built-in defaults: `.ignore`, `.git`, `.gitignore`, `README.*`, `LICENSE.*`, `RCS`, `CVS`
2. A `.ignore` file in the root of `<src>`
3. The `-i` flag

## License

[MIT](LICENSE)

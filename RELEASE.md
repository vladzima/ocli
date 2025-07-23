# Release Process

## Manual Release

1. Update version in `main.go`:
   ```go
   const Version = "1.0.1"
   ```

2. Update CHANGELOG.md with new features/fixes

3. Build release binaries:
   ```bash
   make release
   ```

4. Create Git tag:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```

5. Create GitHub release with binaries from `builds/` directory

## Automated Release (GitHub Actions)

For future automation, create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - name: Build
      run: make release
    - name: Release
      uses: softprops/action-gh-release@v1
      with:
        files: builds/*
```

## Distribution

The package can be installed via:
- `go install github.com/vladarbatov/ocli@latest`
- Direct binary download from releases
- Package managers (future: homebrew, apt, etc.)
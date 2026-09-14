# github-action-pin

[![GitHub Action](https://img.shields.io/github/actions/workflow/status/raviqqe/gh-action-pin/test.yaml?branch=main&style=flat-square)](https://github.com/raviqqe/gh-action-pin/actions)
[![Codecov](https://img.shields.io/codecov/c/github/raviqqe/gh-action-pin.svg?style=flat-square)](https://codecov.io/gh/raviqqe/gh-action-pin)
[![License](https://img.shields.io/github/license/raviqqe/gh-action-pin.svg?style=flat-square)](https://github.com/raviqqe/gh-action-pin/blob/main/UNLICENSE)

Pin GitHub Actions with full semantic versions.

It pins all actions' versions in the Dependabot compatible format at:

- `.github/actions/*/*.yaml`
- `.github/workflows/*/*.yaml`
- `action.yaml` at the repository root

Actions without semantic version tags are pinned to commit hashes of their current references instead.

## Usage

```sh
gh-action-pin
```

For more information, run `gh-action-pin -help`.

## License

[The Unlicense](UNLICENSE)
